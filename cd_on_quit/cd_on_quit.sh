gfm() {
    GFM_LAST_DIR="$(command gfm path-list --lastdir-file 2>/dev/null)"

    if [[ -z "$GFM_LAST_DIR" ]]; then
        os=$(uname -s)
        if [[ "$os" == "Linux" ]]; then
            GFM_LAST_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/t4gfm/lastdir"
        elif [[ "$os" == "Darwin" ]]; then
            GFM_LAST_DIR="$HOME/Library/Application Support/t4gfm/lastdir"
        fi
    fi

    export GFM_LAST_DIR

    command gfm "$@"

	[ ! -f "$GFM_LAST_DIR" ] || {
		. "$GFM_LAST_DIR"
		rm -f -- "$GFM_LAST_DIR" > /dev/null
	}
}
