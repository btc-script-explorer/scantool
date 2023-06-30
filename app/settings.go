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

const TEST_NONE   = byte (0x00)
const TEST_SAVE   = byte (0x01)
const TEST_VERIFY = byte (0x02)
type TestSettings struct {
	testType byte
	directory string
	sourceFile string
}
func NewTestSettings (testType byte, directory string, sourceFile string) TestSettings {
	return TestSettings { testType: testType, directory: directory, sourceFile: sourceFile }
}

func (t *TestSettings) ExitOnError () {

	// make sure the user has the correct permissions
	if t.testType == TEST_SAVE {
		if !checkFile (t.sourceFile, PERM_READ) {
			fmt.Println ("Can not access " + t.sourceFile + ".")
			os.Exit (1)
		}
		if !checkFile (t.directory, PERM_WRITE) {
			fmt.Println ("Can not access " + t.directory + ".")
			os.Exit (1)
		}
	} else if t.testType == TEST_VERIFY {
		if !checkFile (t.directory, PERM_READ) {
			fmt.Println ("Can not access " + t.directory + ".")
			os.Exit (1)
		}
	}
}

func (t *TestSettings) GetTestMode () byte {
	return t.testType
}

func (t *TestSettings) GetDirectory () string {
	return t.directory
}

func (t *TestSettings) GetSourceFile () string {
	return t.sourceFile
}

////////////////////////////////////////////////////////

type AppSettings struct {
	configFile string
}
func NewAppSettings (configFile string) AppSettings {
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
//	nodeLine := "*  Node: " + s.Node.nodeType + " (" + s.Node.GetFullUrl () + ")  "
	nodeLine := "*  Node: " + s.Node.GetFullUrl () + " (" + s.Node.nodeType + ")  "
	webLine := "*   Web: " + s.Website.GetFullUrl () + "  "

	// calculate the width of the message and add padding as necessary
	bannerWidth := len (nodeLine) + 1
	if len (webLine) >= bannerWidth {
		bannerWidth = len (webLine) + 1

		padLen := bannerWidth - len (nodeLine)
		for a := 1; a < padLen; a++ {
			nodeLine += " "
		}
	} else {
		padLen := bannerWidth - len (webLine)
		for a := 1; a < padLen; a++ {
			webLine += " "
		}
	}

	nodeLine += "*"
	webLine += "*"

	// create the lines of the message
	topAndBottom := ""
	for a := 0; a < bannerWidth; a++ {
		topAndBottom += "*"
	}

	fmt.Println ()
	fmt.Println (topAndBottom)
	fmt.Println (nodeLine)
	fmt.Println (webLine)
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

		// website
		"web-url": &settings.Website.url }

	// add parameters that are only allowed on the command line
	stringParamsCommandLineOnly := map [string] *string {
		// app
		"app-config-file": &settings.App.configFile,

		// test
		"test-save-dir": &settings.Test.directory,
		"test-tx-file": &settings.Test.sourceFile,
		"test-verify-dir": &settings.Test.directory }

	if fromCommandLine {
		// add the command line only string type settings
		for paramName, strPointer := range stringParamsCommandLineOnly {
			stringParams [paramName] = strPointer
		}

		// handle parameters that are determined by other settings
		if paramList ["test-save-dir"] != "" {
			settings.Test.testType = TEST_SAVE
		} else if paramList ["test-verify-dir"] != "" {
			settings.Test.testType = TEST_VERIFY
		}
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
	app := NewAppSettings ("")
	node := NewNodeSettings ("BitcoinCore", "127.0.0.1", uint16 (8332), "", "")
	website := NewWebsiteSettings ("127.0.0.1", uint16 (8080))
	test := NewTestSettings (TEST_NONE, "", "./test-transactions.txt")

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

