package scenario

const shellLogger = `# Sourced by Bash startup files and BASH_ENV.
if [[ ${BASH_VERSINFO[0]:-0} -lt 5 || -z ${LB_RUN_ROOT:-} || ${_LB_CAPTURE_PID:-} == "$BASHPID" ]]; then
  return 0
fi
if ! command -v tee >/dev/null 2>&1; then return 0; fi
if ! tee --output-error=warn-nopipe </dev/null >/dev/null 2>&1; then return 0; fi
if ! mkdir -p -- "$LB_RUN_ROOT/shell-logs" 2>/dev/null; then return 0; fi
_LB_CAPTURE_PID=$BASHPID
_LB_CAPTURE_BASE="$LB_RUN_ROOT/shell-logs/bash-$BASHPID-$EPOCHREALTIME"
_lb_capture_record() {
  { printf 'LB-SHELL-1\t%s\t%s\t%s\n' "$EPOCHREALTIME" "$1" "$2"; } 2>/dev/null >> "$3" || :
  return 0
}
_lb_capture_stream() {
  local line
  while IFS= read -r line || [[ -n $line ]]; do
    _lb_capture_record "$1" "$line" "$2"
  done
  return 0
}
if [[ ${0##*/} == collect.sh ]]; then
  { (set -C; printf 'LB-COLLECTION-1\t%s\n' "$EPOCHREALTIME" > "$LB_RUN_ROOT/shell-logs/collection-start"); } 2>/dev/null || :
fi
exec {_LB_CAPTURE_OUT}>&1 {_LB_CAPTURE_ERR}>&2
exec > >(tee --output-error=warn-nopipe >(_lb_capture_stream stdout "$_LB_CAPTURE_BASE.stdout.lb-log") >&"$_LB_CAPTURE_OUT" 2>/dev/null)
exec 2> >(tee --output-error=warn-nopipe >(_lb_capture_stream stderr "$_LB_CAPTURE_BASE.stderr.lb-log") >&"$_LB_CAPTURE_ERR" 2>/dev/null)
if [[ $- != *x* ]]; then
  exec {_LB_CAPTURE_TRACE}> >(_lb_capture_stream command "$_LB_CAPTURE_BASE.commands.lb-log")
  BASH_XTRACEFD=$_LB_CAPTURE_TRACE
  PS4='+ '
  set -x
fi
`
