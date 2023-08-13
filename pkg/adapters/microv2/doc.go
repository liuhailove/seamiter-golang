/*
This package provides sea integration for go-micro.

For server side, users may append a sea handler wrapper to go-micro service, like:

		import (
			seaPlugin "github.com/sea-go/pkg/adapters/micro"
		)

		// Append a sea handler wrapper.
		micro.NewService(micro.WrapHandler(seaPlugin.NewHandlerWrapper()))

The plugin extracts service method as the resource name by default.
Users may provide customized resource name extractor when creating new
sea handler wrapper (via options).

Fallback logic: the plugin will return the BlockError by default
if current request is blocked by sea rules. Users may also
provide customized fallback logic via WithXxxBlockFallback(handler) options.
*/
package micro
