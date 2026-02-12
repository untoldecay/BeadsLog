#!/bin/bash
# monitor.sh - CATCH THE REVERTER

TARGET_FILE="cmd/bd/eval_task_opencode.go"
LOG_FILE="/tmp/git-reverter.log"

echo "🚨 WATCHDOG ACTIVE - Monitoring $TARGET_FILE" | tee -a $LOG_FILE
BASE_HASH=$(md5 -q "$TARGET_FILE")
echo "$(date): BASE_HASH=$BASE_HASH" | tee -a $LOG_FILE

while true; do
  sleep 0.1
  CURRENT_HASH=$(md5 -q "$TARGET_FILE")

  if [ "$CURRENT_HASH" != "$BASE_HASH" ]; then
    echo "💥 REGRESSION DETECTED at $(date)!" | tee -a $LOG_FILE
    echo "Old: $BASE_HASH → New: $CURRENT_HASH" | tee -a $LOG_FILE

    # CULPRIT HUNT
    echo "=== PROCESSES ===" | tee -a $LOG_FILE
    ps aux | grep -E "(git|opencode|bd|beads|fsevent)" | tee -a $LOG_FILE

    echo "=== LOCKS ON FILE ===" | tee -a $LOG_FILE
    lsof "$TARGET_FILE" 2>/dev/null | tee -a $LOG_FILE

    echo "=== GIT STATUS ===" | tee -a $LOG_FILE
    git status --porcelain | tee -a $LOG_FILE

    echo "=== GIT PROCESSES ===" | tee -a $LOG_FILE
    ps aux | grep git | grep -v grep | tee -a $LOG_FILE

    echo "=== FSEVENTS ===" | tee -a $LOG_FILE
    sudo fs_usage -f filesys 2>&1 | grep "$TARGET_FILE" | tail -5 | tee -a $LOG_FILE

    # STOP EVERYTHING
    pkill -f git
    break
  fi
done
