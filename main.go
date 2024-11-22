package main

import (
	"bufio"
	"fmt"
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

