# allows cmd substitution in prompt
setopt prompt_subst

# ─────────────────────────────────────────────────────────────
# Lambda prompt character
# ─────────────────────────────────────────────────────────────
PROMPT_CHAR=";"

# Define Nerd Font symbols matching starship/p10k standards
GIT_ICON=$'\uf1d3'       # Git logo
GITHUB_ICON=$'\uf408'    # GitHub logo
BRANCH_ICON=$'\ue725'        # Git branch icon (alternative)
AHEAD_ICON=$'⇡'          # Up arrow (ahead) - starship standard
BEHIND_ICON=$'⇣'         # Down arrow (behind) - starship standard
DIVERGED_ICON=$'⇕'       # Up/down arrows (diverged) - starship standard
STAGED_ICON=$'+'         # Plus symbol (staged)
UNSTAGED_ICON=$'!'       # Exclamation symbol (unstaged)
UNTRACKED_ICON=$'?'      # Question mark symbol (untracked)
STASHED_ICON=$'$'        # Archive/box symbol (stashed)


# ─────────────────────────────────────────────────────────────
# Git info function with Nerd Font icons and sync status
# ─────────────────────────────────────────────────────────────
git_prompt_info() {
  git rev-parse --git-dir > /dev/null 2>&1 || return

  local branch=$(git symbolic-ref --short HEAD 2>/dev/null || git describe --tags --exact-match 2>/dev/null || echo "DETACHED")
  
  # Single git status call with branch info
  local git_status_raw=$(git status --porcelain -b 2>/dev/null)
  
  # Parse status more efficiently
  local staged=0 unstaged=0 untracked=0
  while IFS= read -r line; do
    case "${line:0:2}" in
      "##") continue ;;  # Branch info line
      [AMDRCU]?) ((staged++)) ;;
      ?[MD]) ((unstaged++)) ;;
      "??") ((untracked++)) ;;
    esac
  done <<< "$git_status_raw"
  
  local stashed=$(git stash list 2>/dev/null | wc -l | tr -d ' ')

  # Detect GitHub by remote URL
  local remote_url=$(git config --get remote.origin.url 2>/dev/null)
  local upstream_url=$(git config --get remote.upstream.url 2>/dev/null)
  
  # Default Git icon
  local host_icon="$GIT_ICON"
  
  # GitHub icon if applicable
  [[ "$remote_url" == *github.com* || "$upstream_url" == *github.com* ]] && host_icon="$GITHUB_ICON"

  # Determine sync (ahead/behind) info - removed sync icon for up-to-date repos
  local ahead behind sync_status=""
  if git rev-parse --abbrev-ref @{u} > /dev/null 2>&1; then
    local counts=$(git rev-list --left-right --count HEAD...@{u} 2>/dev/null)
    ahead=$(echo $counts | awk '{print $1}')
    behind=$(echo $counts | awk '{print $2}')
    
    if [[ $ahead -gt 0 && $behind -gt 0 ]]; then
      sync_status="${DIVERGED_ICON}${ahead}/${behind}"
    elif [[ $ahead -gt 0 ]]; then
      sync_status="${AHEAD_ICON}${ahead}"
    elif [[ $behind -gt 0 ]]; then
      sync_status="${BEHIND_ICON}${behind}"
    fi
  fi

  # Prepare status indicators with icons (no spaces between icon and number)
  local staged_status=""
  local unstaged_status=""
  local untracked_status=""
  local stashed_status=""
  
  [[ $staged -gt 0 ]] && staged_status="${STAGED_ICON}${staged}"
  [[ $unstaged -gt 0 ]] && unstaged_status="${UNSTAGED_ICON}${unstaged}"
  [[ $untracked -gt 0 ]] && untracked_status="${UNTRACKED_ICON}${untracked}"
  [[ $stashed -gt 0 ]] && stashed_status="${STASHED_ICON}${stashed}"

  # Return the git segment info for use in the prompt
  echo "${host_icon}|${BRANCH_ICON}|${branch}|${sync_status}|${staged_status}|${unstaged_status}|${untracked_status}|${stashed_status}"
}


# ─────────────────────────────────────────────────────────────
# Update prompt and Git info
# ─────────────────────────────────────────────────────────────
precmd() {
  export GIT_INFO=$(git_prompt_info)
  
  # Build the prompt
  build_prompt
}

# ─────────────────────────────────────────────────────────────
# Build the prompt with segments
# ─────────────────────────────────────────────────────────────
build_prompt() {
  # Start with a newline
  PROMPT=$'\n'
  
  # Directory segment with melange warm color (#C1A78E)
  PROMPT+="%F{#C1A78E}%~%f"
  
  # Git segment if available
  if [[ -n "$GIT_INFO" ]]; then
    local git_parts=(${(s:|:)GIT_INFO})
    local host_icon=${git_parts[1]}
    local branch_icon=${git_parts[2]}
    local branch=${git_parts[3]}
    local sync_status=${git_parts[4]}
    local staged_status=${git_parts[5]}
    local unstaged_status=${git_parts[6]}
    local untracked_status=${git_parts[7]}
    local stashed_status=${git_parts[8]}
    
    # Add separator and start git segment with melange colors
    PROMPT+=" %F{#867462}on%f %F{#89B3B6}${host_icon} ${branch_icon} %F{#A3A9CE}${branch}%f"
    
    # Add sync status with melange cyan
    if [[ -n "$sync_status" ]]; then
      PROMPT+=" %F{#89B3B6}${sync_status}%f"
    fi
    
    # Add status indicators with melange colors
    if [[ -n "$staged_status" ]]; then
      PROMPT+=" %F{#85B695}${staged_status}%f"  # Green
    fi
    
    if [[ -n "$unstaged_status" ]]; then
      PROMPT+=" %F{#EBC06D}${unstaged_status}%f"  # Yellow
    fi
    
    if [[ -n "$untracked_status" ]]; then
      PROMPT+=" %F{#D47766}${untracked_status}%f"  # Red
    fi
    
    if [[ -n "$stashed_status" ]]; then
      PROMPT+=" %F{#CF9BC2}${stashed_status}%f"  # Purple
    fi
  fi
  
  
  # Add newline and lambda prompt character with melange blue
  PROMPT+=$'\n'"%F{#89B3B6}${PROMPT_CHAR}%f "
}

# Initialize the prompt
build_prompt

