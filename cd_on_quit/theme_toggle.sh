gfm_default() {
    command gfm --theme catppuccin-mocha "$@"
}

gfm_gruvbox() {
    command gfm --theme gruvbox-dark-hard "$@"
}

gfm_toggle() {
    state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/t4gfm"
    state_file="$state_dir/theme-mode"
    mkdir -p "$state_dir"

    mode="default"
    if [ -f "$state_file" ]; then
        mode="$(cat "$state_file" 2>/dev/null)"
    fi

    if [ "$mode" = "gruvbox" ]; then
        printf '%s' "default" > "$state_file"
        command gfm --theme catppuccin-mocha "$@"
        return
    fi

    printf '%s' "gruvbox" > "$state_file"
    command gfm --theme gruvbox-dark-hard "$@"
}
