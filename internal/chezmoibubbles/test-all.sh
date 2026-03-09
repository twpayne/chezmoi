#!/bin/bash

set -eufo pipefail

go tool chezmoi internal-test prompt-bool "yes or no"
go tool chezmoi internal-test prompt-bool overwrite false
go tool chezmoi internal-test prompt-choice color red,green,blue
go tool chezmoi internal-test prompt-choice season spring,summer,fall,winter summer
go tool chezmoi internal-test prompt-int count
go tool chezmoi internal-test prompt-int retries 3
go tool chezmoi internal-test prompt-multichoice directions north,east,south,west
go tool chezmoi internal-test prompt-multichoice days mon,tue,wed,thu,fri,sat,sun sat,sun
go tool chezmoi internal-test prompt-string name
go tool chezmoi internal-test prompt-string username root
go tool chezmoi internal-test read-password
