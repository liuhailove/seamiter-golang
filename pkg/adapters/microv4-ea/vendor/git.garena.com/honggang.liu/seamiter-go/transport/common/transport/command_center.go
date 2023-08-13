package transport

type CommandCenter interface {

	// BeforeStart
	// * Prepare and init for the command center (e.g. register commands).
	// * This will be executed before starting.
	// *
	// * @throws Exception if error occurs
	BeforeStart() error

	// Start
	// * Start the command center in the background.
	// * This method should NOT block.
	// *
	// * @throws Exception if error occurs
	Start() error

	// Stop
	// * Stop the command center and do cleanup.
	// *
	// * @throws Exception if error occurs
	Stop() error
}
