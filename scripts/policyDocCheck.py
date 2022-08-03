#!/usr/bin/env python3

# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
'''Check that policy documentation in a given file in up to date.
The tool checks given documentation file against generated documentation.
'''
import sys

_MARKER = "<!-- BEGIN POLICY-DOC -->"

def check(documentation, generated):
    start = 0
    docLines = documentation.splitlines()
    genLines = generated.splitlines()
    for i, line in enumerate(docLines):
        if line == _MARKER:
            start = i +1
            break

    docLines = docLines[start:]
    if len(docLines) != len(genLines):
        return False, "number of lines in policy documentation is {}, want {}".format(len(docLines), len(genLines))

    for i,line in enumerate(docLines):
        if line != genLines[i]:
            return False, "line {} does not match".format(i)

    return True, None

def main(files):
    errors = []
    try:
      doc = open(files[0]).read()
      generated = open(files[1]).read()
    except Exception as err:
        errors.append(err)

    if errors:
      print('Errors when reading files:')
      print('\n'.join(' - {}'.format(s) for s in errors))
      sys.exit(1)

    valid, err = check(doc, generated)
    if not valid:
      print("Policy documentation differs from generated one, please update:")
      print(' - {}'.format(err))
      sys.exit(1)

if __name__ == '__main__':
  if len(sys.argv) < 2:
    raise SystemExit('Use policyDocCheck.py <docFile> <generatedDocFile>')
  main(sys.argv[1:])
