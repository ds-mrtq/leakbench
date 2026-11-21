from pathlib import Path

settings = dict(line.split('=', 1) for line in Path('.env').read_text().splitlines()
                if line and not line.startswith('#'))
print('Mode: ' + settings['APP_MODE'])
print('HTTP port: ' + settings['HTTP_PORT'])
print('Log level: ' + settings['LOG_LEVEL'])
print('Database and session configuration loaded')
