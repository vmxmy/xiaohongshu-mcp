module.exports = {
  apps: [{
    name: 'xiaohongshu-mcp',
    script: './xiaohongshu-mcp-linux-amd64',
    cwd: '/home/dev/app/xiaohongshu-mcp',
    interpreter: 'none',
    exec_mode: 'fork',
    autorestart: true,
    max_restarts: 10,
    restart_delay: 5000
  }]
}
