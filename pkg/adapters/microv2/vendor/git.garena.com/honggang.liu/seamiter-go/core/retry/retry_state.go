package retry

// RtyState Stateful retry is characterised by having to recognise the items that are
// being processed, so this interface is used primarily to provide a cache key in
// between failed attempts. It also provides a hints to the
// {@link RtyOperations} for optimisations to do with avoidable cache hits and
// switching to stateless retry if a rollback is not needed.
type RtyState interface {

	// GetKey
	// Key representing the state for a retry attempt. Stateful retry is
	// characterised by having to recognise the items that are being processed,
	// so this value is used as a cache key in between failed attempts.
	//
	// @return the key that this state represents
	//
	GetKey() interface{}

	// IsForceRefresh
	// Indicate whether a cache lookup can be avoided. If the key is known ahead
	// of the retry attempt to be fresh (i.e. has never been seen before) then a
	// cache lookup can be avoided if this flag is true.
	//
	// @return true if the state does not require an explicit check for the key
	IsForceRefresh() bool

	// RollbackFor
	// Check whether this exception requires a rollback. The default is always
	// true, which is conservative, so this method provides an optimisation for
	// switching to stateless retry if there is an exception for which rollback
	// is unnecessary. Example usage would be for a stateful retry to specify a
	// validation exception as not for rollback.
	//
	// @param exception the exception that caused a retry attempt to fail
	// @return true if this exception should cause a rollback
	RollbackFor(err error) bool
}
