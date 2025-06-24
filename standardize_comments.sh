#!/bin/bash

# Script pour standardiser tous les commentaires GoDoc

find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Possible error/error if any/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Manager name/name: string package manager name/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Shell command string/cmd: string shell command string/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Command for installed version/installed: string shell command for installed version/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Command for available version/available: string shell command for available version/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Error if no package is specified/err: error if no package is specified/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Error if package name is empty/err: error if package name is empty/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Error if any/err: error if any/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Context for timeout and cancellation/ctx: context.Context for timeout and cancellation/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/List of packages to install/pkgs: ...string list of packages to install/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/List of packages to uninstall/pkgs: ...string list of packages to uninstall/g' {} \;
find . -name "*.go" -not -name "*_test.go" -exec sed -i '' 's/Package name to check/pkg: string package name to check/g' {} \;

echo "Standardisation terminée !"
