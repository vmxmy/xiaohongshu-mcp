module.exports = {
  apps: [
    {
      name: 'xiaohongshu-mcp',
      script: './xiaohongshu-mcp',
      args: '--headless=true --port=:18060',
      cwd: '/path/to/xiaohongshu-mcp', // 修改为实际部署路径

      // 实例配置
      instances: 1,
      exec_mode: 'fork',

      // 自动重启
      autorestart: true,
      watch: false,
      max_memory_restart: '500M',

      // 环境变量
      env: {
        NODE_ENV: 'production',
        PORT: '18060',
        // 如果需要指定浏览器路径
        // ROD_BROWSER_BIN: '/usr/bin/chromium-browser'
      },

      // 日志配置
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      error_file: './logs/mcp-error.log',
      out_file: './logs/mcp-out.log',
      log_file: './logs/mcp-combined.log',

      // 日志轮转
      max_restarts: 10,
      min_uptime: '10s',

      // 进程ID文件
      pid_file: './pids/mcp.pid',

      // 监听文件变化（开发模式可开启）
      ignore_watch: [
        'node_modules',
        'logs',
        'pids',
        '.git',
        'cookies.json',
        'pic'
      ],

      // 优雅退出
      kill_timeout: 5000,
      wait_ready: true,
      listen_timeout: 10000,

      // 错误处理
      merge_logs: true,
      time: true
    }
  ]
}
