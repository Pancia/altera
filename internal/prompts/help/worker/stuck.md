# Worker: When Stuck

If you're blocked or confused:

1. **Write a help message** to checkpoint.md:
   ```
   echo "STUCK: <description of problem>" > checkpoint.md
   ```
2. **Signal the daemon** with `alt checkpoint <your-agent-id>`
3. **Include context** in your checkpoint:
   - What you tried
   - What error or confusion you hit
   - What you think the problem might be
4. **Wait for guidance** — the liaison or human will respond

Common reasons for getting stuck:
- Task description is ambiguous → ask for clarification via checkpoint
- Dependencies not met → check if blocking tasks are complete
- Tests fail in unexpected ways → document the failures
- Missing access or permissions → signal immediately
