package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ConnectorStyle struct{
	Intermediate	 string
	Last 			 string
	Prefix			 string
	Branch 			 string
}

func patternToRegex(pattern string) (string,error){
	regexPattern:= regexp.QuoteMeta(pattern)

	regexPattern=strings.ReplaceAll(regexPattern,`\*`,`.*`)
	regexPattern=strings.ReplaceAll(regexPattern,`\?`,`.`)

	regexPattern ="^" +regexPattern + "$"

	_,err:= regexp.Compile(regexPattern)
	if err!=nil{
		return "",fmt.Errorf("Error in compiling the regex pattern: %s",err)
	}

	return regexPattern,nil
}

func scanDirectory(root string,prefix string,ignoredDirs map[string]struct{},style ConnectorStyle,excludePatterns []string,maxDepth,currentDepth int) (string,error){
	logrus.Debugf("Scanning directory: %s with prefix %s",root,prefix)
	var result strings.Builder
	entries,err:= os.ReadDir(root)
	if err!=nil{
		return "",fmt.Errorf("Error in reading directory: %s",err)
	}

	filteredEntries:=[]os.DirEntry{}
	for _,entry:=range entries{
		if _,ok :=ignoredDirs[entry.Name()];ok{
			logrus.Debugf("Ignoring directory: %s",entry.Name())
			continue
		}

		excluded :=false
		for _,pattern:=range excludePatterns{
			regexPattern,err:= patternToRegex(pattern)
			if err!=nil {
				return "",fmt.Errorf("Error in converting pattern %s to regex: %v",pattern,err)
			}

			matched,err:=regexp.MatchString(regexPattern,entry.Name())
			if err!=nil{
				return "",fmt.Errorf("Error in matching pattern %s with entry: %s",pattern,entry.Name())
			}

			if matched{
				logrus.Debugf("Excluding directory: %s",entry.Name())
				excluded=true
				break
			}
		}

		if !excluded{
			filteredEntries=append(filteredEntries,entry)
		}
	}

	if maxDepth!=-1 && currentDepth>maxDepth{
		return "",nil
	}

	for i,entry:=range filteredEntries{
		connector := style.Intermediate
		newPrefix:= prefix+style.Branch
		if i==len(filteredEntries)-1{
			connector=style.Last
			newPrefix=prefix+style.Prefix
		}

		result.WriteString((fmt.Sprintf("%s%s%s\n",prefix,connector,entry.Name())))

		if entry.IsDir(){
			subDir,err:= scanDirectory(filepath.Join(root,entry.Name()),newPrefix,ignoredDirs,style,excludePatterns,maxDepth,currentDepth+1)
			if err!=nil{
				return "",err
			}
			result.WriteString(subDir)
		}
	}
	return result.String(),nil
}

func readDirIgnore(root string)(map[string]struct{},error){
	ignoredDirs:= make(map[string]struct{})
	dirIgnorePath:= filepath.Join(root,".dirignore")

	file,err:= os.Open(dirIgnorePath)
	if err!=nil{
		if os.IsNotExist(err){
			return ignoredDirs,nil
		}
		return nil,fmt.Errorf("Error in opening file: %s",err)
	}
	defer file.Close()

	scanner:=bufio.NewScanner(file)
	for scanner.Scan(){
		dir:=strings.TrimSpace(scanner.Text())
		if dir!=""{
			ignoredDirs[dir]=struct{}{}
		}
	}

	if err:= scanner.Err();err!=nil{
		return nil,fmt.Errorf("Error in reading .dirignore: %v",err)
	}

	return ignoredDirs,nil
}

func main(){
	var(
		intermediate string
		last string
		prefix string
		branch string
		exclude []string
		maxDepth int
		verbose bool
	)

	var rootCmd = &cobra.Command{}


	rootCmd.Flags().StringVar(&intermediate, "intermediate", "├── ", "Symbol for intermediate nodes")
		rootCmd.Flags().StringVar(&last, "last", "└── ", "Symbol for the last node in a directory")
		rootCmd.Flags().StringVar(&prefix, "prefix", "    ", "Prefix for child nodes")
		rootCmd.Flags().StringVar(&branch, "branch", "│   ", "Branch for intermediate nodes")
		rootCmd.Flags().StringSliceVar(&exclude, "exclude", []string{}, "Exclude files or directories matching these patterns (e.g., '*.txt')")
		rootCmd.Flags().IntVar(&maxDepth, "depth", -1, "Limit the depth of the directory traversal (-1 for no limit)")
		rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Whether or not to show debug messages")

		if err := rootCmd.Execute(); err != nil {
			logrus.Fatal(err)
		}
}
