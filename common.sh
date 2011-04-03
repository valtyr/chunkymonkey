# Ensures the build environment is clean for either building or testing.
function clean_for () {
    desired_state=$1

    state_file=".last_build_type"
    last=`cat $state_file 2>/dev/null || echo dirty`
    if [ "x$last" != "x$desired_state" ]; then
        echo "Cleaning up after last build type: $last"
        gd -c src/lib

        # Clean up any old test .go files.
        if [ "x$last" = "xtest" ]; then
            rm -r src/lib/tmp*
        fi
    fi
    echo $desired_state > $state_file
}
