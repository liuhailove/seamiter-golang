package file

import (
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type RefreshableFileDataSource struct {
	datasource.Base
	sourceFilePath string
	sourceFileName string
	isInitialized  util.AtomicBool
	closeChan      chan struct{}
	watcher        *fsnotify.Watcher
	closed         util.AtomicBool
}

func (s *RefreshableFileDataSource) Write(bytes []byte) error {
	f, err := os.OpenFile(filepath.Join(s.sourceFilePath, s.sourceFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Errorf("RefreshableFileDataSource fail to open the property file, err: %+v.", err)
	}
	defer f.Close()
	_, err = f.Write(bytes)
	if err != nil {
		logging.Error(err, "RefreshableFileDataSource fail to write the property file, err", err)
		return errors.Errorf("RefreshableFileDataSource fail to write the property file, err: %+v.", err)
	}
	return nil
}

func NewFileDataSource(sourceFilePath, sourceFileName string, handlers ...datasource.PropertyHandler) *RefreshableFileDataSource {
	var ds = &RefreshableFileDataSource{
		sourceFilePath: sourceFilePath,
		sourceFileName: sourceFileName,
		closeChan:      make(chan struct{}),
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds
}
func (s *RefreshableFileDataSource) ReadSource() ([]byte, error) {
	f, err := os.Open(filepath.Join(s.sourceFilePath, s.sourceFileName))
	if err != nil {
		return nil, errors.Errorf("RefreshableFileDataSource fail to open the property file, err: %+v.", err)
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Errorf("RefreshableFileDataSource fail to read file, err: %+v.", err)
	}
	return src, nil
}
func (s *RefreshableFileDataSource) Initialize() error {
	if !s.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	exist, err := util.FileExists(filepath.Join(s.sourceFilePath, s.sourceFileName))
	if err != nil {
		logging.Error(err, "Fail to execute RefreshableFileDataSource.FileExists")
	}
	if !exist {
		err = util.CreateDirIfNotExists(s.sourceFilePath)
		if err != nil {
			logging.Error(err, "Fail to execute RefreshableFileDataSource.CreateDirIfNotExists")
			return err
		}
		_, err = os.OpenFile(filepath.Join(s.sourceFilePath, s.sourceFileName), os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			logging.Error(err, "Fail to execute RefreshableFileDataSource.WriteFile")
			return err
		}
	}
	err = s.doReadAndUpdate()
	if err != nil {
		logging.Error(err, "Fail to execute RefreshableFileDataSource.doReadAndUpdate")
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Errorf("Fail to new a watcher instance of fsnotify, err: %+v", err)
	}
	err = w.Add(filepath.Join(s.sourceFilePath, s.sourceFileName))
	if err != nil {
		return errors.Errorf("Fail add a watcher on file[%s], err: %+v", s.sourceFilePath, err)
	}
	s.watcher = w

	go util.RunWithRecover(func() {
		defer s.watcher.Close()
		for {
			select {
			case ev := <-s.watcher.Events:
				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					logging.Warn("[RefreshableFileDataSource] The file source was renamed.", "sourceFilePath", s.sourceFilePath)
					updateErr := s.Handle(nil)
					if updateErr != nil {
						logging.Error(updateErr, "Fail to update nil property")
					}

					// try to watch sourceFile
					_ = s.watcher.Remove(filepath.Join(s.sourceFilePath, s.sourceFileName))
					retryCount := 0
					for {
						if retryCount > 5 {
							logging.Error(errors.New("retry failed"), "Fail to retry watch", "sourceFilePath", s.sourceFilePath)
							s.Close()
							return
						}
						e := s.watcher.Add(filepath.Join(s.sourceFilePath, s.sourceFileName))
						if e == nil {
							break
						}
						retryCount++
						logging.Error(e, "Failed to add to watcher", "sourceFilePath", s.sourceFilePath)
						util.Sleep(time.Second)
					}
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					logging.Warn("[RefreshableFileDataSource] The file source was removed.", "sourceFilePath", s.sourceFilePath)
					updateErr := s.Handle(nil)
					if updateErr != nil {
						logging.Error(updateErr, "Fail to update nil property")
					}
					s.Close()
					return
				}
				err := s.doReadAndUpdate()
				if err != nil {
					logging.Error(err, "Fail to execute RefreshableFileDataSource.doReadAndUpdate")
				}
			case err := <-s.watcher.Errors:
				logging.Error(err, "Watch err on file", "sourceFilePath", s.sourceFilePath)
			case <-s.closeChan:
				return
			}
		}
	})
	return nil
}

func (s *RefreshableFileDataSource) doReadAndUpdate() (err error) {
	src, err := s.ReadSource()
	if err != nil {
		err = errors.Errorf("Failed to read source,err: %+v", err)
		return err
	}
	return s.Handle(src)
}

func (s *RefreshableFileDataSource) Close() error {
	if !s.closed.CompareAndSet(false, true) {
		return nil
	}
	s.closeChan <- struct{}{}
	logging.Info("[File] The RefreshableFileDataSource for file had been closed.", "sourceFilePath", s.sourceFilePath, "sourceFileName", s.sourceFileName)
	return nil
}
