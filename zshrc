
# allows cmd substitution in prompt
setopt prompt_subst

# ─────────────────────────────────────────────────────────────
# Powerlevel10k style prompt character
# ─────────────────────────────────────────────────────────────
PROMPT_CHAR="❯"

# Define Nerd Font symbols with explicit Unicode values
GIT_ICON=$'\uf1d3'       # Git logo
GITHUB_ICON=$'\uf408'    # GitHub logo
BRANCH_ICON=$'\uf418'    # Branch icon
AHEAD_ICON=$'\uf176'     # Arrow up (ahead)
BEHIND_ICON=$'\uf175'    # Arrow down (behind)
DIVERGED_ICON=$'\uf7a5'  # Up/down arrows (diverged)
SYNC_ICON=$'\u2714'      # Checkmark/tick symbol (✔)
STAGED_ICON=$'\uf055'    # Plus symbol (staged)
UNSTAGED_ICON=$'\uf06a'  # Exclamation symbol (unstaged)
UNTRACKED_ICON=$'\uf29c' # Question mark symbol (untracked)
STASHED_ICON=$'\uf01c'   # Archive/box symbol (stashed)

# Powerline rounded edge symbols
ROUNDED_LEFT=$'\uE0B6'   # Rounded left edge
ROUNDED_RIGHT=$'\uE0B4'  # Rounded right edge

# ─────────────────────────────────────────────────────────────
# Git info function with Nerd Font icons and sync status
# ─────────────────────────────────────────────────────────────
git_prompt_info() {
  git rev-parse --git-dir > /dev/null 2>&1 || return

  local branch=$(git symbolic-ref --short HEAD 2>/dev/null || git describe --tags --exact-match 2>/dev/null)
  [[ -z "$branch" ]] && branch=" DETACHED"

  local git_status_raw=$(git status --porcelain 2>/dev/null)
  local staged=$(echo "$git_status_raw" | grep -E '^[AMDRCU]' | wc -l | tr -d ' ')
  local unstaged=$(echo "$git_status_raw" | grep -E '^.[MD]' | wc -l | tr -d ' ')
  local untracked=$(echo "$git_status_raw" | grep -E '^\?\?' | wc -l | tr -d ' ')
  local stashed=$(git stash list 2>/dev/null | wc -l | tr -d ' ')

  # Detect GitHub by remote URL
  local remote_url=$(git config --get remote.origin.url 2>/dev/null)
  local upstream_url=$(git config --get remote.upstream.url 2>/dev/null)
  
  # Default Git icon
  local host_icon="$GIT_ICON"
  
  # GitHub icon if applicable
  [[ "$remote_url" == *github.com* || "$upstream_url" == *github.com* ]] && host_icon="$GITHUB_ICON"
  
  local git_bg_color="green"
  local git_fg_color="black"
  
  if [[ $((staged + unstaged + untracked + stashed)) -gt 0 ]]; then
    git_bg_color="red"
  fi

  # Determine sync (ahead/behind) info
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
    else
      sync_status="${SYNC_ICON}"
    fi
  fi

  # Prepare status indicators with icons
  local staged_status=""
  local unstaged_status=""
  local untracked_status=""
  local stashed_status=""
  
  [[ $staged -gt 0 ]] && staged_status="${STAGED_ICON} ${staged}"
  [[ $unstaged -gt 0 ]] && unstaged_status="${UNSTAGED_ICON} ${unstaged}"
  [[ $untracked -gt 0 ]] && untracked_status="${UNTRACKED_ICON} ${untracked}"
  [[ $stashed -gt 0 ]] && stashed_status="${STASHED_ICON} ${stashed}"

  # Return the git segment info for use in the prompt
  echo "${git_bg_color}|${git_fg_color}|${host_icon}|${BRANCH_ICON}|${branch}|${sync_status}|${staged_status}|${unstaged_status}|${untracked_status}|${stashed_status}"
}

# ─────────────────────────────────────────────────────────────
# Update prompt timestamp and Git info
# ─────────────────────────────────────────────────────────────
precmd() {
  export GIT_INFO=$(git_prompt_info)
  export PROMPT_TIMESTAMP=$(date +"%H:%M:%S")
  
  # Build the prompt
  build_prompt
}

# ─────────────────────────────────────────────────────────────
# Build the prompt with segments
# ─────────────────────────────────────────────────────────────
build_prompt() {
  # Start with a newline
  PROMPT=$'\n'
  
  # Add rounded left edge with cyan color
  PROMPT+="%F{cyan}${ROUNDED_LEFT}"
  
  # First segment (directory)
  PROMPT+="%K{cyan}%F{black} %~ %f%k"
  
  # Git segment if available (no space before it)
  if [[ -n "$GIT_INFO" ]]; then
    local git_parts=(${(s:|:)GIT_INFO})
    local git_bg=${git_parts[1]}
    local git_fg=${git_parts[2]}
    local host_icon=${git_parts[3]}
    local branch_icon=${git_parts[4]}
    local branch=${git_parts[5]}
    local sync_status=${git_parts[6]}
    local staged_status=${git_parts[7]}
    local unstaged_status=${git_parts[8]}
    local untracked_status=${git_parts[9]}
    local stashed_status=${git_parts[10]}
    
    # Start git segment
    PROMPT+="%K{$git_bg}%F{$git_fg}"
    
    # Add git icons with the same color as the text
    PROMPT+=" ${host_icon} ${branch_icon} ${branch}"
    
    # Add sync status if available
    [[ -n "$sync_status" ]] && PROMPT+=" ${sync_status}"
    
    # Add status indicators if available
    [[ -n "$staged_status" ]] && PROMPT+=" ${staged_status}"
    [[ -n "$unstaged_status" ]] && PROMPT+=" ${unstaged_status}"
    [[ -n "$untracked_status" ]] && PROMPT+=" ${untracked_status}"
    [[ -n "$stashed_status" ]] && PROMPT+=" ${stashed_status}"
    
    # End git segment
    PROMPT+=" %k"
    
    # Add timestamp segment directly after git segment
    PROMPT+="%K{240}%F{white} ${PROMPT_TIMESTAMP} "
  else
    # If no git info, add timestamp directly after directory
    PROMPT+="%K{240}%F{white} ${PROMPT_TIMESTAMP} "
  fi
  
  # Add rounded right edge with gray color
  PROMPT+="%f%k%F{240}${ROUNDED_RIGHT}"
  
  # Add newline and Powerlevel10k style prompt character for the second line (no background)
  PROMPT+=$'\n'"%F{blue}${PROMPT_CHAR}%f "
}

# Initialize the prompt
build_prompt

