package app

import (
	"fmt"
	"os"
	"io/fs"
	"bufio"
	"strings"
	"strconv"
	"sync"

//	"btctx/btc"
)

type NodeSettings struct {
	nodeType string
	url string
	port uint16
	username string
	password string
}
func NewNodeSettings (nodeType string, url string, port uint16, username string, password string) NodeSettings {
	return NodeSettings { nodeType: nodeType, url: url, port: port, username: username, password: password }
}

func (n *NodeSettings) GetNodeType () string {
	return n.nodeType
}

func (n *NodeSettings) GetUrl () string {
	return n.url
}

func (n *NodeSettings) GetFullUrl () string {
	return n.url + ":" + strconv.FormatUint (uint64 (n.port), 10)
}

func (n *NodeSettings) GetPort () uint16 {
	return n.port
}

func (n *NodeSettings) GetUsername () string {
	return n.username
}

func (n *NodeSettings) GetPassword () string {
	return n.password
}

////////////////////////////////////////////////////////

type WebsiteSettings struct {
	url string
	port uint16
}
func NewWebsiteSettings (url string, port uint16) WebsiteSettings {
	return WebsiteSettings { url: url, port: port }
}

func (w *WebsiteSettings) ExitOnError () {

	// TODO: make sure the url is a valid interface and the port is not already open
	// if not implemented these errors will show up when http.ListenAndServe is called

}

func (w *WebsiteSettings) GetUrl () string {
	return w.url
}

func (w *WebsiteSettings) GetFullUrl () string {
	return w.url + ":" + strconv.FormatUint (uint64 (w.port), 10)
}

func (w *WebsiteSettings) GetPort () uint16 {
	return w.port
}

////////////////////////////////////////////////////////

type TestSettings struct {
	testMode string
	verifiedDir string
	unverifiedDir string
	sourceFile string
}
func NewTestSettings (testMode string, verifiedDir string, unverifiedDir string, sourceFile string) TestSettings {
	mode := strings.ToLower (testMode)
	if mode != "" && mode != "save" && mode != "verify" {
		fmt.Println (mode + " is not a valid test mode.")
		os.Exit (1)
	}
	return TestSettings { testMode: mode, verifiedDir: verifiedDir, unverifiedDir: unverifiedDir, sourceFile: sourceFile }
}

func (t *TestSettings) ExitOnError () {

	// make sure the user has the correct permissions
	if t.testMode == "save" {
		if !checkFile (t.sourceFile, PERM_READ) {
			fmt.Println ("Can not access " + t.sourceFile + ".")
			os.Exit (1)
		}
		if !checkFile (t.unverifiedDir, PERM_WRITE) {
			fmt.Println ("Can not access " + t.unverifiedDir + ".")
			os.Exit (1)
		}
	} else if t.testMode == "verify" {
		if !checkFile (t.verifiedDir, PERM_READ) {
			fmt.Println ("Can not access " + t.verifiedDir + ".")
			os.Exit (1)
		}
	}
}

func (t *TestSettings) GetTestMode () string {
	return t.testMode
}

func (t *TestSettings) GetDirectory () string {
	if t.testMode == "" {
		return ""
	} else if t.testMode == "save" {
		return t.unverifiedDir
	}
	return t.verifiedDir
}

func (t *TestSettings) GetSourceFile () string {
	return t.sourceFile
}

////////////////////////////////////////////////////////

type AppSettings struct {
	configFile string
}
func NewAppSettings (configFile string, themeName string, layoutName string) AppSettings {
	return AppSettings { configFile: configFile }
}

func (a *AppSettings) GetConfigFileLocation () string {
	return a.configFile
}

////////////////////////////////////////////////////////

type settingsManager struct {
	Node NodeSettings
	Website WebsiteSettings
	App AppSettings
	Test TestSettings
}

// singleton
var settings *settingsManager = nil
var once sync.Once

func GetSettings () *settingsManager {
	once.Do (parseSettings)
	return settings
}

////////////////////////////////////////////////////////

const PERM_READ  = byte (0b00000100)
const PERM_WRITE = byte (0b00000010)
const PERM_EXEC  = byte (0b00000001)
func checkFile (fileName string, requiredPermissions byte) bool {

	// make sure the file or directory exists
	fileInfo, err := os.Stat (fileName)
	if err != nil {
		fmt.Println (err.Error ())
		return false
	}

	// make sure the user has the correct permissions
	hasPermission := true
	if requiredPermissions != 0 {
		hasPermission = fileInfo.Mode ().Perm () & fs.FileMode (requiredPermissions << 8) != 0 ||
						fileInfo.Mode ().Perm () & fs.FileMode (requiredPermissions << 4) != 0 ||
						fileInfo.Mode ().Perm () & fs.FileMode (requiredPermissions) != 0
	}
	if !hasPermission {
		return false
	}

	return true
}

func (s *settingsManager) PrintListeningMessage () {

	// create the data lines of the message
	lines := make ([] string, 2)
	lines [0] = "*    Node: " + s.Node.GetFullUrl () + " (" + s.Node.nodeType + ")  "
	lines [1] = "*     Web: " + s.Website.GetFullUrl () + "  "

	// calculate the width of the message and add padding as necessary
	bannerWidth := len (lines [0]) + 1
	for l := 1; l < len (lines); l++ {
		if len (lines [l]) + 1 > bannerWidth {
			bannerWidth = len (lines [l]) + 1
		}
	}

	topAndBottom := ""
	for a := 0; a < bannerWidth; a++ {
		topAndBottom += "*"
	}

	// pad the ones that need to be padded
	for l := 0; l < len (lines); l++ {
		if len (lines [l]) + 1 < bannerWidth {
			padLen := bannerWidth - len (lines [l])
			for a := 1; a < padLen; a++ {
				lines [l] += " "
			}
		}

		lines [l] += "*"
	}

	// print the message
	fmt.Println ()
	fmt.Println (topAndBottom)
	for l := 0; l < len (lines); l++ {
		fmt.Println (lines [l])
	}
	fmt.Println (topAndBottom)
	fmt.Println ()
}

func (s *settingsManager) parseParamList (paramList map [string] string, fromCommandLine bool) {

	// create the map of string type parameters
	stringParams := map [string] *string {
		// node
		"node-type": &settings.Node.nodeType,
		"node-url": &settings.Node.url,
		"node-user": &settings.Node.username,
		"node-pw": &settings.Node.password,

		// test
		"test-save-dir": &settings.Test.unverifiedDir,
		"test-tx-file": &settings.Test.sourceFile,
		"test-verify-dir": &settings.Test.verifiedDir,

		// website
		"web-url": &settings.Website.url }

	// create the map of parameters that are only allowed on the command line
	stringParamsCommandLineOnly := map [string] *string {
		// app
		"app-config-file": &settings.App.configFile,

		// test
		"test-mode": &settings.Test.testMode }

	if fromCommandLine {
		// add the command line only string type settings
		for paramName, strPointer := range stringParamsCommandLineOnly {
			stringParams [paramName] = strPointer
		}
	}

	// handle parameters that can be determined by other settings
	if paramList ["test-save-dir"] != "" {
		settings.Test.testMode = "save"
	} else if paramList ["test-verify-dir"] != "" {
		settings.Test.testMode = "verify"
	}

	// create the map of uint16 type parameters
	uint16Params := map [string] *uint16 {
		// node
		"node-port": &settings.Node.port,

		// website
		"web-port": &settings.Website.port }

	for paramName, valueStr := range paramList {
		if stringParams [paramName] != nil {
			*stringParams [paramName] = valueStr
		} else if uint16Params [paramName] != nil {
			v, err := strconv.Atoi (valueStr)
			if err != nil {
				fmt.Println ("Error parsing " + paramName + ": " + err.Error ())
				os.Exit (1)
			}
			*uint16Params [paramName] = uint16 (v)
		} else if stringParamsCommandLineOnly [paramName] != nil && !fromCommandLine {
			fmt.Println (paramName + " is ignored in the config file, only supported from the command line.")
		} else {
			fmt.Println (paramName + " is not a recognized setting.")
		}
	}
}

func parseSettings () {

	if os.Args [0] == "--help" {
		// TODO: print help message
		os.Exit (0)
	}

	// default settings
	app := NewAppSettings ("", "Default", "Desktop")
	node := NewNodeSettings ("BitcoinCore", "127.0.0.1", uint16 (8332), "", "")
	website := NewWebsiteSettings ("127.0.0.1", uint16 (8080))
	test := NewTestSettings ("", "./tests/verified-transactions/", "./tests/unverified-transactions/", "./tests/transactions.txt")

	settings = &settingsManager { Node: node, Website: website, App: app, Test: test }

	// create the map of command line parameters
	commandLineParameters := make (map [string] string)
	commandLineParamCount := len (os.Args)
	for a := 1; a < commandLineParamCount; a++ {

		// remove the -- from the front of the parameter
		parameter := os.Args [a]
		if len (parameter) < 2 || parameter [0:2] != "--" {
			fmt.Println (parameter + " is improperly formatted.")
			continue
		}

		// add the parameter to the map
		parts := strings.Split (parameter [2:], "=")
		if len (parts) != 2 {
			fmt.Println (parameter + " is improperly formatted.")
			continue
		}

		commandLineParameters [parts [0]] = parts [1]
	}

	// make sure there aren't any conflicting command line parameters
	if len (commandLineParameters ["test-save-dir"]) > 0 && len (commandLineParameters ["test-verify-dir"]) > 0 {
		fmt.Println ("Parameters test-save-dir and test-verify-dir can not both be set.")
		os.Exit (1)
	}

	// check for config file parameters
	configFileParameters := make (map [string] string)
	configFileLocation, configFileProvided := commandLineParameters ["app-config-file"]
	if configFileProvided {
		configFile, err := os.Open (configFileLocation)
		if err != nil {
			fmt.Println (err.Error ())
		}
		fileScanner := bufio.NewScanner (configFile)
		for fileScanner.Scan () {
			parameterStr := strings.TrimSpace (fileScanner.Text ())

			// skip comments and blank lines
			if parameterStr == "" || parameterStr [0] == '#' {
				continue
			}

			// if there are any spaces on the line, only use the first string of text and make sure the rest is a comment
			separateStrings := strings.Split (parameterStr, " ")
			separateStringCount := len (separateStrings)
			if separateStringCount > 1 {
				// get the next bit of contiguous text from this line
				nextString := ""
				for s := 1; s < separateStringCount; s++ {
					if len (separateStrings [s]) > 0 {
						nextString = separateStrings [s]
						break
					}
				}

				if nextString [0] == '#' {
					// it's a comment, so the line of text is okay so far
					parameterStr = separateStrings [0]
				} else {
					// it is not a comment, we assume the line is an error
					fmt.Println ("Improperly-formatted line in config file: " + parameterStr)
					continue
				}
			}

			// get the parameter and its string value
			parts := strings.Split (parameterStr, "=")
			if len (parts) != 2 {
				fmt.Println (parameterStr + " is not a properly-formatted setting.")
				continue
			}

			// the command line takes precedence over the config file
			if commandLineParameters [parts [0]] == "" {
				configFileParameters [parts [0]] = parts [1]
			} else {
				fmt.Println (parts [0] + " setting was provided more than once. Command line value of " + commandLineParameters [parts [0]] + " will be used.")
			}
		}
	}

	// update the settings from the command line and config file
	settings.parseParamList (commandLineParameters, true)
	if len (configFileParameters) > 0 {
		settings.parseParamList (configFileParameters, false)
	}
}

