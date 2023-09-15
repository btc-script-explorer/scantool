package app

import (
	"fmt"
	"os"
	"io/fs"
	"io/ioutil"
	"bufio"
	"strings"
	"strconv"
)

type settingsManager struct {

	alreadyParsed bool
	versionTag string

	configFile string

	bitcoinCoreAddr string
	bitcoinCorePort uint16
	bitcoinCoreUsername string
	bitcoinCorePassword string

	nodeVersionStr string

	addr string
	port uint16

	noWeb bool

	testMode string
	testVerifiedDir string
	testUnverifiedDir string
	testSourceFile string
}

func (s *settingsManager) ExitOnError () {

	// verify the web settings
	if s.noWeb && (len (s.addr) == 0 || s.port == 0) { panic ("Web parameters are not valid.") }

	// make sure the user has the correct test file permissions
	if s.testMode == "save" {
		if !checkFile (s.testSourceFile, PERM_READ) { panic ("Can not access test souce file " + s.testSourceFile + ".") }
		if !checkFile (s.testUnverifiedDir, PERM_WRITE) { panic ("Can not access unverified test directory " + s.testUnverifiedDir + ".") }
	} else if s.testMode == "verify" {
		if !checkFile (s.testVerifiedDir, PERM_READ) { panic ("Can not access verified test directory " + s.testVerifiedDir + ".") }
	}
}

func (s *settingsManager) GetConfigFileLocation () string {
	return s.configFile
}

func (s *settingsManager) GetNodeType () string {
	if len (s.bitcoinCoreAddr) > 0 && s.bitcoinCorePort != 0 && len (s.bitcoinCoreUsername) > 0 && len (s.bitcoinCorePassword) > 0 {
		return "Bitcoin Core"
	}

	return ""
}

func (s *settingsManager) GetNodeFullUrl () string {
	return s.bitcoinCoreAddr + ":" + strconv.FormatUint (uint64 (s.bitcoinCorePort), 10)
}

func (s *settingsManager) GetNodeUsername () string {
	return s.bitcoinCoreUsername
}

func (s *settingsManager) GetNodePassword () string {
	return s.bitcoinCorePassword
}

func (s *settingsManager) GetBaseUrl (alwaysIncludePort bool) string {
	if s.port != 80 {
		return fmt.Sprintf ("%s:%d", s.addr, s.port)
	}

	return s.addr
}

func (s *settingsManager) GetFullUrl () string {
	return fmt.Sprintf ("http://%s", s.GetBaseUrl (false))
}

func (s *settingsManager) GetAddr () string {
	return s.addr
}

func (s *settingsManager) GetPort () uint16 {
	return s.port
}

func (s *settingsManager) GetTestMode () string {
	return s.testMode
}

func (s *settingsManager) GetTestDirectory () string {
	if s.testMode == "save" {
		return s.testUnverifiedDir
	} else if s.testMode == "verify" {
		return s.testVerifiedDir
	}
	return ""
}

func (s *settingsManager) GetTestSourceFile () string {
	return s.testSourceFile
}

func (s *settingsManager) IsWebOn () bool {
	return !s.noWeb
}

func (s *settingsManager) setSettings (settings map [string] string) {
	for k, v := range settings {
		switch k {
			case "config-file": s.configFile = v

			// bitcoin core
			case "bitcoin-core-addr": s.bitcoinCoreAddr = v
			case "bitcoin-core-port":
				port, err := strconv.Atoi (v)
				if err != nil { panic (err.Error ()) }
				s.bitcoinCorePort = uint16 (port)
			case "bitcoin-core-username": s.bitcoinCoreUsername = v
			case "bitcoin-core-password": s.bitcoinCorePassword = v

			// web
			case "addr": s.addr = v
			case "port":
				port, err := strconv.Atoi (v)
				if err != nil { panic (err.Error ()) }
				s.port = uint16 (port)
			case "no-web": s.noWeb = true

			// test
			case "test-mode": s.testMode = v
			case "test-verified-dir": s.testVerifiedDir = v
			case "test-unverified-dir": s.testUnverifiedDir = v
			case "test-source-file": s.testSourceFile = v
		}
	}
}

var Settings settingsManager

func getDefaultSettings () settingsManager {
	return settingsManager {
//								configFile: "",

								bitcoinCoreAddr: "127.0.0.1",
								bitcoinCorePort: 8332,
//								bitcoinCoreUsername: "",
//								bitcoinCorePassword: "",

								addr: "127.0.0.1",
								port: 8080,
//								noWeb: false,

//								testMode: "",
//								testVerifiedDir: "",
//								testUnverifiedDir: "",
//								testSourceFile: ""
							}
}

func ParseSettings () {
	if Settings.alreadyParsed { return }

	Settings = getDefaultSettings ()

	versionBytes, err := ioutil.ReadFile ("./VERSION")
	if err != nil { fmt.Println (err.Error ()) }
	Settings.versionTag = string (versionBytes)
	for Settings.versionTag [len (Settings.versionTag) - 1] == '\n' {
		Settings.versionTag = Settings.versionTag [0 : len (Settings.versionTag) - 1]
	}

	parameters := make (map [string] string)

	// command line parameters
	commandLineParamCount := len (os.Args)
	for a := 1; a < commandLineParamCount; a++ {

		// remove the -- from the front of the parameter
		parameter := os.Args [a]
		if len (parameter) < 2 || parameter [0:2] != "--" { panic (parameter + " is improperly formatted.") }

		// add the parameter to the map
		parts := strings.Split (parameter [2:], "=")
		if len (parts) != 2 { panic (parameter + " parameter is improperly formatted.") }

		if parameters [parts [0]] == "" {
			parameters [parts [0]] = parts [1]
		}
	}

	// config file parameters
	if len (parameters ["config-file"]) > 0 {
		configFileLines := readConfigFile (parameters ["config-file"])

		for _, line := range configFileLines {

			// get the parameter and its string value
			parts := strings.Split (line, "=")
			key := parts [0]
			value := ""; if len (parts) >= 2 { value = parts [1] }

			if parameters [key] == "" {
				parameters [key] = value
			}
		}
	}

	Settings.setSettings (parameters)

	Settings.ExitOnError ()
	Settings.alreadyParsed = true
}

func readConfigFile (configFileLocation string) [] string {

	configFile, err := os.Open (configFileLocation)
	if err != nil { panic (err.Error ()) }

	var configFileLines [] string

	fileScanner := bufio.NewScanner (configFile)
	for fileScanner.Scan () {
		parameterStr := strings.TrimSpace (fileScanner.Text ())

		// skip comments and blank lines
		if len (parameterStr) == 0 || parameterStr [0] == '#' { continue }

		// if there are any spaces on the line, only use the first string of text and make sure the rest is a comment
		separateStrings := strings.Split (parameterStr, " ")
		separateStringCount := len (separateStrings)
		if separateStringCount > 1 {
			// get the next bit of contiguous text from this line
			// this is done in a loop because consecutive spaces can return empty strings
			nextString := ""
			for s := 1; s < separateStringCount; s++ {
				if len (separateStrings [s]) > 0 {
					nextString = separateStrings [s]
					break
				}
			}

			// if it is a comment, the line of text is okay
			// it is not a comment, we assume the line is an error
			if nextString [0] == '#' {
				parameterStr = separateStrings [0]
			} else {
				panic ("Improperly-formatted line in config file: " + parameterStr)
			}
		}

		configFileLines = append (configFileLines, parameterStr)
	}

	configFile.Close ()
	return configFileLines
}

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

func (s *settingsManager) SetNodeVersionString (nodeVersionStr string) {
	s.nodeVersionStr = nodeVersionStr
}

func (s *settingsManager) PrintListeningMessage () {

	// create the data lines of the message
	lines := make ([] string, 0)
	lines = append (lines, "")
	lines = append (lines, "SCANTOOL " + GetVersion ())
	lines = append (lines, "")
	lines = append (lines, s.nodeVersionStr)
	lines = append (lines, s.GetNodeFullUrl ())
	lines = append (lines, "")
	lines = append (lines, "Web Access:")
	lines = append (lines, s.GetFullUrl () + "/web/")
	lines = append (lines, "")
	lines = append (lines, "REST Example:")
	lines = append (lines, "curl -X GET " + s.GetFullUrl () + "/rest/v1/current_block_height")
	lines = append (lines, "")

	// calculate the width of the message and add padding as necessary
	bannerWidth := 0
	for l := 0; l < len (lines); l++ {
		if len (lines [l]) % 2 != 0 {
			lines [l] += " "
		}

		if len (lines [l]) + 6 > bannerWidth {
			bannerWidth = len (lines [l]) + 6
		}
	}

	topAndBottom := ""
	for a := 0; a < bannerWidth; a++ {
		topAndBottom += "*"
	}

	// pad the ones that need to be padded
	for l := 0; l < len (lines); l++ {
		for len (lines [l]) < bannerWidth - 2 {
			lines [l] = " " + lines [l] + " "
		}
	}

	// print the message
	fmt.Println ()
	fmt.Println (topAndBottom)
	for l := 0; l < len (lines); l++ {
		fmt.Println ("*" + lines [l] + "*")
	}
	fmt.Println (topAndBottom)
	fmt.Println ()
}

func GetVersion () string {
	return Settings.versionTag
}
