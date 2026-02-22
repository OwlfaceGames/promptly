# melange.promptly.fish
# Warm color palette inspired by the Melange Neovim theme

# ─────────────────────────────────────────────────────────────
# Melange palette
# ─────────────────────────────────────────────────────────────
set -g __melange_dir      C1A78E   # warm sand — directory
set -g __melange_muted    867462   # warm gray — "on" label
set -g __melange_cyan     89B3B6   # muted teal — git icons, sync, prompt char
set -g __melange_purple   A3A9CE   # soft lavender — branch name
set -g __melange_green    85B695   # sage green — staged
set -g __melange_yellow   EBC06D   # warm amber — unstaged
set -g __melange_red      D47766   # terracotta — untracked
set -g __melange_accent   CF9BC2   # mauve — stashed

# ─────────────────────────────────────────────────────────────
# Icons (Nerd Font — matching melange.promptly.zsh)
# ─────────────────────────────────────────────────────────────
set -g __melange_github_icon  \uf408   # GitHub logo
set -g __melange_git_icon     \uf1d3   # Git logo
set -g __melange_branch_icon  \ue725   # branch
set -g __melange_ahead        ⇡
set -g __melange_behind       ⇣
set -g __melange_diverged     ⇕
set -g __melange_prompt_char  ";"

# ─────────────────────────────────────────────────────────────
# Helper — set_color wrapper for hex values
# ─────────────────────────────────────────────────────────────
function __mel
    set_color $argv[1]
    echo -n $argv[2]
    set_color normal
end

# ─────────────────────────────────────────────────────────────
# Git segment
# ─────────────────────────────────────────────────────────────
function __melange_git_segment
    # Bail if not in a git repo
    git rev-parse --git-dir >/dev/null 2>&1; or return

    set -l branch (git symbolic-ref --short HEAD 2>/dev/null; \
                   or git describe --tags --exact-match 2>/dev/null; \
                   or echo DETACHED)

    # Detect GitHub remote
    set -l remote_url (git config --get remote.origin.url 2>/dev/null)
    set -l host_icon $__melange_git_icon
    if string match -q "*github.com*" $remote_url
        set host_icon $__melange_github_icon
    end

    # Status counts
    set -l staged    (git diff --cached --name-only 2>/dev/null | count)
    set -l unstaged  (git diff --name-only 2>/dev/null | count)
    set -l untracked (git ls-files --others --exclude-standard 2>/dev/null | count)
    set -l stashed   (git stash list 2>/dev/null | count)

    # Ahead / behind
    set -l sync_status ""
    if git rev-parse --abbrev-ref @{u} >/dev/null 2>&1
        set -l counts (git rev-list --left-right --count HEAD...@{u} 2>/dev/null)
        set -l ahead_n  (echo $counts | awk '{print $1}')
        set -l behind_n (echo $counts | awk '{print $2}')
        if test $ahead_n -gt 0 -a $behind_n -gt 0
            set sync_status "$__melange_diverged$ahead_n/$behind_n"
        else if test $ahead_n -gt 0
            set sync_status "$__melange_ahead$ahead_n"
        else if test $behind_n -gt 0
            set sync_status "$__melange_behind$behind_n"
        end
    end

    # Print segment
    echo -n " "
    __mel $__melange_muted "on"
    echo -n " "
    __mel $__melange_cyan "$host_icon $__melange_branch_icon "
    __mel $__melange_purple $branch

    test -n "$sync_status"   && echo -n " " && __mel $__melange_cyan    $sync_status
    test $staged    -gt 0    && echo -n " " && __mel $__melange_green   "+$staged"
    test $unstaged  -gt 0    && echo -n " " && __mel $__melange_yellow  "!$unstaged"
    test $untracked -gt 0    && echo -n " " && __mel $__melange_red     "?$untracked"
    test $stashed   -gt 0    && echo -n " " && __mel $__melange_accent  "\$$stashed"
end

# ─────────────────────────────────────────────────────────────
# Main prompt function (fish auto-loads this from functions/)
# ─────────────────────────────────────────────────────────────
function fish_prompt
    echo ""   # leading newline

    # Directory
    __mel $__melange_dir (prompt_pwd)

    # Git segment (only if inside a repo)
    __melange_git_segment

    # Newline + prompt character
    echo ""
    __mel $__melange_cyan $__melange_prompt_char
    echo -n " "
end
