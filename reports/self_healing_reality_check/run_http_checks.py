"""Local source-only acceptance driver; never runs downloaded code."""
import json, pathlib, time, urllib.request, urllib.error, subprocess

OUT = pathlib.Path(__file__).parent / 'evidence'
BASE = 'http://127.0.0.1:18080/api'
trace = []

def request(path, body=None):
    start = time.monotonic()
    req = urllib.request.Request(BASE + path, data=None if body is None else json.dumps(body).encode(),
        headers={'X-API-Token': 'reality-local-test-token', 'Content-Type': 'application/json'})
    try:
        with urllib.request.urlopen(req, timeout=180) as response:
            code, raw = response.status, response.read().decode()
    except urllib.error.HTTPError as error:
        code, raw = error.code, error.read().decode()
    try: result = json.loads(raw)
    except ValueError: result = {'raw': raw}
    trace.append({'path':path, 'input':body, 'status':code, 'output':result, 'duration_seconds':time.monotonic()-start})
    (OUT / 'http-trace.json').write_text(json.dumps(trace, indent=2))
    print(path, code, flush=True)
    return code, result

if __name__ == '__main__':
    request('/scan', {})
    request('/scan', {'repo_url':'not-a-url','retain_workspace':True})
    request('/remediation/targets')
    request('/remediation/targets?workspace_id=wrong-workspace')
    request('/remediation/analyze', {'target_id':'INJ-EXT2-RES-02'})
    request('/remediation/execute', {'token_id':'forged'})
    request('/scan', {'repo_url':'https://github.com/Zylus08/reality-check-does-not-exist-20261125','retain_workspace':True})
    for owner, name in [('Zylus08','mlsc-timetable'),('go-chi','chi'),('google','uuid'),('Zylus08','Finger-Counter')]:
        code, result = request('/scan', {'repo_url':f'https://github.com/{owner}/{name}', 'retain_workspace':True})
        if code != 200: continue
        ws = result['workspace_id']
        root = pathlib.Path('E:/SWE/go-backend/uploads') / ('workspace_' + ws)
        sha = subprocess.check_output(['git','-C',str(root),'rev-parse','HEAD'],text=True).strip()
        (OUT / (name+'-http-revision.json')).write_text(json.dumps({'repository':owner+'/'+name,'commit':sha,'workspace_id':ws},indent=2))
        request('/remediation/targets?workspace_id='+ws)
