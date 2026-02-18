gfm() {
    os=$(uname -s)

	# Linux
	if [[ "$os" == "Linux" ]]; then
		export GFM_LAST_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/t4gfm/lastdir"
	fi

	# macOS
	if [[ "$os" == "Darwin" ]]; then
		export GFM_LAST_DIR="$HOME/Library/Application Support/t4gfm/lastdir"
	fi

    command gfm "$@"

	[ ! -f "$GFM_LAST_DIR" ] || {
		. "$GFM_LAST_DIR"
		rm -f -- "$GFM_LAST_DIR" > /dev/null
	}
}
