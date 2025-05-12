# allows cmd substitution in prompt
setopt prompt_subst

# ─────────────────────────────────────────────────────────────
# Git info function with Nerd Font icons and sync status
# ─────────────────────────────────────────────────────────────
git_prompt_info() {
  git rev-parse --git-dir > /dev/null 2>&1 || return

  local branch=$(git symbolic-ref --short HEAD 2>/dev/null || git describe --tags --exact-match 2>/dev/null)
  [[ -z "$branch" ]] && branch=" DETACHED"

  local git_status_raw=$(git status --porcelain 2>/dev/null)
  local staged=$(echo "$git_status_raw" | grep -E '^[AMDRCU]' | wc -l | tr -d ' ')
  local unstaged=$(echo "$git_status_raw" | grep -E '^.[MD]' | wc -l | tr -d ' ')
  local untracked=$(echo "$git_status_raw" | grep -E '^\?\?' | wc -l | tr -d ' ')
  local stashed=$(git stash list 2>/dev/null | wc -l | tr -d ' ')

  # Detect GitHub by remote URL
  local remote_url=$(git config --get remote.origin.url 2>/dev/null)
  local upstream_url=$(git config --get remote.upstream.url 2>/dev/null)
  local host_icon=""  # default Git icon
  [[ "$remote_url" == *github.com* || "$upstream_url" == *github.com* ]] && host_icon=""  # GitHub icon

  local branch_icon=""
  local git_color="%F{green}"
  [[ $((staged + unstaged + untracked + stashed)) -gt 0 ]] && git_color="%F{red}"

  # Determine sync (ahead/behind) info
  local ahead behind sync_icon=""
  if git rev-parse --abbrev-ref @{u} > /dev/null 2>&1; then
    local counts=$(git rev-list --left-right --count HEAD...@{u} 2>/dev/null)
    ahead=$(echo $counts | awk '{print $1}')
    behind=$(echo $counts | awk '{print $2}')
    if [[ $ahead -gt 0 && $behind -gt 0 ]]; then
      sync_icon="⇅$ahead/$behind"
    elif [[ $ahead -gt 0 ]]; then
      sync_icon="↑$ahead"
    elif [[ $behind -gt 0 ]]; then
      sync_icon="↓$behind"
    else
      sync_icon="✓"
    fi
  fi

  # Build final Git segment (without square brackets around sync)
  local out="${git_color}${host_icon} ${branch_icon} $branch"
  [[ -n $sync_icon ]] && out+=" $sync_icon"
  [[ $staged    -gt 0 ]] && out+=" $staged"
  [[ $unstaged  -gt 0 ]] && out+=" $unstaged"
  [[ $untracked -gt 0 ]] && out+=" $untracked"
  [[ $stashed   -gt 0 ]] && out+=" $stashed"
  out+="%f"

  echo "$out"
}

# ─────────────────────────────────────────────────────────────
# Update prompt timestamp and Git info
# ─────────────────────────────────────────────────────────────
precmd() {
  export GIT_INFO=$(git_prompt_info)
  export PROMPT_TIMESTAMP=$(date +"%H:%M:%S")
}

# ─────────────────────────────────────────────────────────────
# Final two-line prompt definition
# ─────────────────────────────────────────────────────────────
PROMPT=$'\n%F{cyan}%~%f${GIT_INFO:+ $GIT_INFO} %F{244}$PROMPT_TIMESTAMP%f\n%F{blue}⚡ %f'

