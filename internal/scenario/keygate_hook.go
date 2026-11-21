package scenario

const keyGateHook = `
# Best-effort Bash pre-command gate; not an OS security boundary.
_KG_EXE=$(command -v kg) || _KG_EXE=
_KG_BUSY=0
_kg_check() {
  local status=$1 text=$2 source=$3 caller=$4 traced=$- result=0 json char escaped code i
  set +x
  if [[ $_KG_BUSY == 1 || -z $_KG_EXE || ! -x $_KG_EXE ]]; then
    [[ $traced != *x* ]] || set -x
    return 0
  fi
  case "$source" in
    */.bashrc|*/.profile|*/.bash_profile|*/.bash_login|*/.session-init|*/env.sh|*/session-init.sh|/etc/profile|/etc/bash.bashrc)
      [[ $traced != *x* ]] || set -x
      return 0 ;;
  esac
  case "$caller" in _sh_trace*|_kg_*)
    [[ $traced != *x* ]] || set -x
    return 0 ;;
  esac
  shopt -s extdebug
  _KG_BUSY=1
  escaped=
  for ((i=0; i<${#text}; i++)); do
    char=${text:i:1}
    case "$char" in
      '"') escaped+='\"' ;;
      '\') escaped+='\\' ;;
      *)
        printf -v code '%d' "'$char"
        if ((code < 32)); then
          printf -v char '\\u%04x' "$code"
        fi
        escaped+=$char ;;
    esac
  done
  printf -v json '{"tool_name":"Bash","tool_input":{"command":"%s"}}' "$escaped"
  if env -u BASH_ENV -u ENV "$_KG_EXE" hook check <<< "$json" >&2; then
    result=0
  else
    result=$?
  fi
  _KG_BUSY=0
  [[ $traced != *x* ]] || set -x
  if [[ $result == 0 ]]; then return 0; fi
  return 1
}
trap '_kg_check "$?" "$BASH_COMMAND" "${BASH_SOURCE[0]-}" "${FUNCNAME[0]-}"' DEBUG
`
