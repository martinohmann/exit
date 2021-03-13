package exit

const (
	// Generic codes.
	CodeOK      = 0 // success
	CodeErr     = 1 // generic error
	CodeHelpErr = 2 // command is invoked with -help or -h flag but no such flag is defined

	// Codes as defined in /usr/include/sysexits.h on *nix systems.
	CodeUsage       = 64 // command line usage error
	CodeDataErr     = 65 // data format error
	CodeNoInput     = 66 // cannot open input
	CodeNoUser      = 67 // addressee unknown
	CodeNoHost      = 68 // host name unknown
	CodeUnavailable = 69 // service unavailable
	CodeSoftware    = 70 // internal software error
	CodeOSErr       = 71 // system error (e.g., can't fork)
	CodeOSFile      = 72 // critical OS file missing
	CodeCantCreat   = 73 // can't create (user) output file
	CodeIOErr       = 74 // input/output error
	CodeTempFail    = 75 // temp failure; user is invited to retry
	CodeProtocol    = 76 // remote error in protocol
	CodeNoPerm      = 77 // permission denied
	CodeConfig      = 78 // configuration error
)
