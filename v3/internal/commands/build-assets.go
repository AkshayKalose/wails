package commands

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/leaanthony/gosod"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed build_assets
var buildAssets embed.FS

//go:embed updatable_build_assets
var updatableBuildAssets embed.FS

type BuildAssetsOptions struct {
	Dir                string `description:"The directory to generate the files into" default:"."`
	Name               string `description:"The name of the project"`
	Binary             string `description:"The name of the binary"`
	ProductName        string `description:"The name of the product" default:"My Product"`
	ProductDescription string `description:"The description of the product" default:"My Product Description"`
	ProductVersion     string `description:"The version of the product" default:"0.1.0"`
	ProductCompany     string `description:"The company of the product" default:"My Company"`
	ProductCopyright   string `description:"The copyright notice" default:"\u00a9 now, My Company"`
	ProductComments    string `description:"Comments to add to the generated files" default:"This is a comment"`
	ProductIdentifier  string `description:"The product identifier, e.g com.mycompany.myproduct"`
	Silent             bool   `description:"Suppress output to console"`
}

type UpdateBuildAssetsOptions struct {
	Dir    string `description:"The directory to generate the files into" default:"build"`
	Config string `description:"The config file to use" default:"appdata.yaml"`
	Silent bool   `description:"Suppress output to console"`
}

type Config struct {
	Binary             string `yaml:"binaryName"`
	ProductCompany     string `yaml:"companyName"`
	ProductName        string `yaml:"productName"`
	ProductIdentifier  string `yaml:"productIdentifier"`
	ProductDescription string `yaml:"description"`
	ProductVersion     string `yaml:"productVersion"`
	ProductCopyright   string `yaml:"copyright"`
	ProductComments    string `yaml:"comments"`
}

func GenerateBuildAssets(options *BuildAssetsOptions) error {
	DisableFooter = true

	var err error
	options.Dir, err = filepath.Abs(options.Dir)
	if err != nil {
		return err
	}

	// If directory doesn't exist, create it
	if _, err := os.Stat(options.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(options.Dir, 0755)
		if err != nil {
			return err
		}
	}

	if options.ProductComments == "" {
		options.ProductComments = fmt.Sprintf("(c) %d %s", time.Now().Year(), options.ProductCompany)
	}

	if options.ProductIdentifier == "" {
		options.ProductIdentifier = "com.wails." + normaliseName(options.Name)
	}

	if options.Binary == "" {
		options.Binary = normaliseName(options.Name)
		if runtime.GOOS == "windows" {
			options.Binary += ".exe"
		}
	}

	tfs, err := fs.Sub(buildAssets, "build_assets")
	if err != nil {
		return err
	}

	if !options.Silent {
		println("Generating build assets in " + options.Dir)
	}
	err = gosod.New(tfs).Extract(options.Dir, options)
	if err != nil {
		return err
	}
	tfs, err = fs.Sub(updatableBuildAssets, "updatable_build_assets")
	if err != nil {
		return err
	}
	return gosod.New(tfs).Extract(options.Dir, options)
}

func UpdateBuildAssets(options *UpdateBuildAssetsOptions) error {
	DisableFooter = true

	if options.Config == "" {
		return fmt.Errorf("config file required for update")
	}

	var err error
	options.Dir, err = filepath.Abs(options.Dir)
	if err != nil {
		return err
	}

	// If directory doesn't exist, create it
	if _, err := os.Stat(options.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(options.Dir, 0755)
		if err != nil {
			return err
		}
	}

	tfs, err := fs.Sub(updatableBuildAssets, "updatable_build_assets")
	if err != nil {
		return err
	}

	configFile := filepath.Join(options.Dir, options.Config)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file %s does not exist", configFile)
	}

	config, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var configData Config
	err = yaml.Unmarshal(config, &configData)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", configData)

	err = gosod.New(tfs).Extract(options.Dir, configData)
	if err != nil {
		return err
	}

	if !options.Silent {
		println("Successfully updated build assets in " + options.Dir)
	}

	return nil
}

func normaliseName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
