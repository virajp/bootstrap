# Set the terminal title to the current git branch if we're in a
# git repository, otherwise show the current directory.
function fish_title
    set -l git_root (git rev-parse --show-toplevel 2>/dev/null)
    if test -n "$git_root"
        set -l branch (git branch --show-current 2>/dev/null)
        if test -n "$branch"
            echo (basename $git_root)" · $branch"
        else
            echo (basename $git_root)
        end
    else
        echo (basename $PWD)
    end
end
