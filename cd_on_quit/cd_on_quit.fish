function gfm
    set os $(uname -s)

    if test "$os" = "Linux"
        set gfm_last_dir "$HOME/.local/state/t4gfm/lastdir"
    end

    if test "$os" = "Darwin"
        set gfm_last_dir "$HOME/Library/Application Support/t4gfm/lastdir"
    end

    command gfm $argv

    if test -f "$gfm_last_dir"
        source "$gfm_last_dir"
        rm -f -- "$gfm_last_dir" >> /dev/null
    end
end
