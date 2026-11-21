package scenario

const shellLogger = `# Sourced by Bash startup files and BASH_ENV.
if [[ ${BASH_VERSINFO[0]:-0} -lt 5 || -z ${SH_RUN_ROOT:-} || ${_SH_TRACE_PID:-} == "$BASHPID" ]]; then
  return 0
fi
if ! command -v tee >/dev/null 2>&1; then return 0; fi
if ! tee --output-error=warn-nopipe </dev/null >/dev/null 2>&1; then return 0; fi
if ! mkdir -p -- "$SH_RUN_ROOT/shell-logs" 2>/dev/null; then return 0; fi
_SH_TRACE_PID=$BASHPID
_SH_TRACE_BASE="$SH_RUN_ROOT/shell-logs/bash-$BASHPID-$EPOCHREALTIME"
_sh_trace_record() {
  { printf 'SH-SESSION-1\t%s\t%s\t%s\n' "$EPOCHREALTIME" "$1" "$2"; } 2>/dev/null >> "$3" || :
  return 0
}
_sh_trace_stream() {
  local line source_text keep=0
  while IFS= read -r line || [[ -n $line ]]; do
    if [[ $1 == command ]]; then
      if [[ $line == +*SH-TRACE$'\t'* ]]; then
        source_text=${line#*SH-TRACE$'\t'}
        source_text=${source_text%%$'\t'*}
        line=${line#*SH-TRACE$'\t'}
        line=${line#*$'\t'}
        keep=1
        case " $source_text " in
          *"/.profile"*|*" .profile "*|\
          *"/.bashrc"*|*" .bashrc "*|\
          *"/.bash_profile"*|*" .bash_profile "*|\
          *"/.bash_login"*|*" .bash_login "*|\
          *"/.session-init"*|*" .session-init "*|\
          *"/env.sh"*|*" env.sh "*|\
          *"/session-init.sh"*|*" session-init.sh "*|\
          *"/etc/profile"*|*"/etc/bash.bashrc"*) keep=0 ;;
        esac
      fi
      [[ $keep == 1 ]] || continue
    fi
    _sh_trace_record "$1" "$line" "$2"
  done
  return 0
}
if [[ ${0##*/} == collect.sh ]]; then
  { (set -C; printf 'SH-COLLECTION-1\t%s\n' "$EPOCHREALTIME" > "$SH_RUN_ROOT/shell-logs/collection-start"); } 2>/dev/null || :
fi
exec {_SH_TRACE_OUT}>&1 {_SH_TRACE_ERR}>&2
exec > >(
  exec 1>&"$_SH_TRACE_OUT" 2>/dev/null
  exec {_SH_TRACE_OUT}>&- {_SH_TRACE_ERR}>&-
  tee --output-error=warn-nopipe >(exec >/dev/null 2>&1; _sh_trace_stream stdout "$_SH_TRACE_BASE.stdout.session-log")
)
exec 2> >(
  exec 1>&"$_SH_TRACE_ERR" 2>/dev/null
  exec {_SH_TRACE_OUT}>&- {_SH_TRACE_ERR}>&-
  tee --output-error=warn-nopipe >(exec >/dev/null 2>&1; _sh_trace_stream stderr "$_SH_TRACE_BASE.stderr.session-log")
)
exec {_SH_TRACE_OUT}>&- {_SH_TRACE_ERR}>&-
if [[ $- != *x* ]]; then
  exec {_SH_TRACE_TRACE}> >(exec >/dev/null 2>&1; _sh_trace_stream command "$_SH_TRACE_BASE.commands.session-log")
  BASH_XTRACEFD=$_SH_TRACE_TRACE
  PS4=$'+SH-TRACE\t${BASH_SOURCE[*]}\t'
  set -x
fi
`
