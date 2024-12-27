/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package utils is related with storage client utils
package utils

import (
	"errors"
	"strings"
	"text/template"

	"github.com/huawei/csm/v2/utils/log"
)

// TextTemplate is for text template
type TextTemplate struct {
	text *template.Template
}

// Format generate a text with args
func (t *TextTemplate) Format(args map[string]interface{}) (string, error) {
	str := new(strings.Builder)
	err := t.text.Execute(str, args)
	if err != nil {
		log.Errorln("format text template error")
		return "", errors.New("format text template error")
	}
	return str.String(), nil
}

// NewTextTemplate is used to create text template
func NewTextTemplate(templateName string, str string) *TextTemplate {
	temp, err := template.New(templateName).Parse(str)
	if err != nil {
		log.Errorln("init text template error, name: %s, str: %s", templateName, str)
		return nil
	}
	return &TextTemplate{text: temp}
}
