import os

mode = os.environ.get('WORKER_MODE', 'batch')
print('Worker mode: ' + mode)
print('Region: ' + os.environ.get('WORKER_REGION', 'ci'))
print('Queued 0 jobs' if mode == 'preview' else 'Queued 3 jobs')
