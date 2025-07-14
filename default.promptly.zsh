# allows cmd substitution in prompt
setopt prompt_subst

# ─────────────────────────────────────────────────────────────
# Lambda prompt character
# ─────────────────────────────────────────────────────────────
PROMPT_CHAR="❯"

# Define Nerd Font symbols with explicit Unicode values
AHEAD_ICON=$'\uf176'     # Arrow up (ahead)
BEHIND_ICON=$'\uf175'    # Arrow down (behind)
DIVERGED_ICON=$'\uf7a5'  # Up/down arrows (diverged)
STAGED_ICON=$'+'         # Plus symbol (staged)
UNSTAGED_ICON=$'!'       # Exclamation symbol (unstaged)
UNTRACKED_ICON=$'?'      # Question mark symbol (untracked)
STASHED_ICON=$'$'        # Archive/box symbol (stashed)

# ─────────────────────────────────────────────────────────────
# Git info function with text labels instead of icons
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
  
  # Default Git text
  local host_text="git"
  
  # GitHub text if applicable
  [[ "$remote_url" == *github.com* || "$upstream_url" == *github.com* ]] && host_text="github"

  # Determine sync (ahead/behind) info - removed sync icon for up-to-date repos
  local ahead behind sync_status=""
  if git rev-parse --abbrev-ref @{u} > /dev/null 2>&1; then
    local counts=$(git rev-list --left-right --count HEAD...@{u} 2>/dev/null)
    ahead=$(echo $counts | awk '{print $1}')
    behind=$(echo $counts | awk '{print $2}')
    
    if [[ $ahead -gt 0 && $behind -gt 0 ]]; then
      sync_status="${DIVERGED_ICON} ${ahead}/${behind}"
    elif [[ $ahead -gt 0 ]]; then
      sync_status="${AHEAD_ICON} ${ahead}"
    elif [[ $behind -gt 0 ]]; then
      sync_status="${BEHIND_ICON} ${behind}"
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

  # Return the git segment info WITH sync status
  echo "${host_text}|${branch}|${sync_status}|${staged_status}|${unstaged_status}|${untracked_status}|${stashed_status}"
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
  
  # Directory segment in cyan
  PROMPT+="%F{cyan}%~%f"
  
  # Git segment if available
  if [[ -n "$GIT_INFO" ]]; then
    local git_parts=(${(s:|:)GIT_INFO})
    local host_text=${git_parts[1]}
    local branch=${git_parts[2]}
    local sync_status=${git_parts[3]}
    local staged_status=${git_parts[4]}
    local unstaged_status=${git_parts[5]}
    local untracked_status=${git_parts[6]}
    local stashed_status=${git_parts[7]}
    
    # Build git segment with explicit color boundaries
    PROMPT+=" %F{blue}${host_text}(%F{013}${branch}%F{blue})%f"
    
    # Add sync status with cyan color (unique from everything else)
    if [[ -n "$sync_status" ]]; then
      PROMPT+=" %F{cyan}${sync_status}%f"
    fi
    
    # Add status indicators with explicit color wrapping
    if [[ -n "$staged_status" ]]; then
      PROMPT+=" %F{green}${staged_status}%f"
    fi
    
    if [[ -n "$unstaged_status" ]]; then
      PROMPT+=" %F{yellow}${unstaged_status}%f"
    fi
    
    if [[ -n "$untracked_status" ]]; then
      PROMPT+=" %F{red}${untracked_status}%f"
    fi
    
    if [[ -n "$stashed_status" ]]; then
      PROMPT+=" %F{white}${stashed_status}%f"
    fi
  fi
  
  # Add newline and lambda prompt character for the second line
  PROMPT+=$'\n'"%F{blue}${PROMPT_CHAR}%f "
}

# Initialize the prompt
build_prompt

