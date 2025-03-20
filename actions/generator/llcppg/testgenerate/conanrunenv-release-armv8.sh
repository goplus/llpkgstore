script_folder="/workspaces/project/llpkgstore/actions/generator/llcppg/testgenerate"
echo "echo Restoring environment" > "$script_folder/deactivate_conanrunenv-release-armv8.sh"
for v in LD_LIBRARY_PATH DYLD_LIBRARY_PATH
do
    is_defined="true"
    value=$(printenv $v) || is_defined="" || true
    if [ -n "$value" ] || [ -n "$is_defined" ]
    then
        echo export "$v='$value'" >> "$script_folder/deactivate_conanrunenv-release-armv8.sh"
    else
        echo unset $v >> "$script_folder/deactivate_conanrunenv-release-armv8.sh"
    fi
done


export LD_LIBRARY_PATH="/home/vscode/.conan2/p/b/cjson020f476755124/p/lib:$LD_LIBRARY_PATH"
export DYLD_LIBRARY_PATH="/home/vscode/.conan2/p/b/cjson020f476755124/p/lib:$DYLD_LIBRARY_PATH"