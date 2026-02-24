import requests
import json
import time
import concurrent.futures

BASE_URL = "http://localhost:8080"
TOKEN_FILE = "token.json"

def get_token():
    with open(TOKEN_FILE, "r") as f:
        data = json.load(f)
        return data["token"]

def submit_and_wait(problem_id, code, description):
    token = get_token()
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    payload = {
        "problem_id": problem_id,
        "source_code": code
    }
    
    # Submit
    resp = requests.post(f"{BASE_URL}/api/submissions", headers=headers, json=payload)
    if not resp.ok:
        print(f"FAILED to submit {description}: {resp.status_code} - {resp.text}")
        return None
    
    submission_id = resp.json()["submission_id"]
    print(f"Submitted {description}, ID: {submission_id}")
    
    # Poll for result (simplifying for test script instead of SSE)
    for _ in range(20):
        time.sleep(1)
        resp = requests.get(f"{BASE_URL}/api/submissions", headers=headers)
        submissions = resp.json()
        for s in submissions:
            if s["ID"] == submission_id:
                if s["Status"] not in ["PENDING", "RUNNING"]:
                    return s
    return None

ADVERSARIAL_TESTS = [
    {
        "name": "File System Access (Read /etc/passwd)",
        "code": "def f(a,b,c):\n    with open('/etc/passwd', 'r') as f: print(f.read())",
        "expected": ["RE", "WA"] # Should fail
    },
    {
        "name": "Network Access (Google)",
        "code": "import urllib.request\ndef f(a,b,c):\n    print(urllib.request.urlopen('http://google.com').read())",
        "expected": ["RE", "TLE"] # Should fail or timeout
    },
    {
        "name": "Memory Bomb (Large Allocation)",
        "code": "def f(a,b,c):\n    x = [0] * (50 * 10**6) # ~400MB to exceed 64MB",
        "expected": ["OLE", "RE"] # Memory Limit Exceeded (137)
    },
    {
        "name": "CPU Bomb (While True)",
        "code": "def f(a,b,c):\n    while True: pass",
        "expected": ["TLE"]
    }
]

def run_adversarial():
    print("--- Starting Adversarial Tests ---")
    for test in ADVERSARIAL_TESTS:
        result = submit_and_wait(1001, test["code"], test["name"])
        if result:
            status = result["Status"]
            msg = result["Message"]
            print(f"RESULT: [{status}] {msg}")
            if status in test["expected"]:
                 print(f"✅ PASSED (Expected failure caught)")
            else:
                 print(f"❌ FAILED (Should have been blocked as {test['expected']})")
        else:
            print(f"❌ TIMEOUT while waiting for result")
    print("")

def run_concurrency(count=10):
    print(f"--- Starting Concurrency Stress Test ({count} concurrent) ---")
    valid_code = "def f(a,b,c): return a+b+c"
    token = get_token()
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    def single_submit(i):
        start = time.time()
        resp = requests.post(f"{BASE_URL}/api/submissions", headers=headers, json={
            "problem_id": 1001,
            "source_code": valid_code
        })
        return resp.ok

    with concurrent.futures.ThreadPoolExecutor(max_workers=count) as executor:
        results = list(executor.map(single_submit, range(count)))
    
    success = sum(results)
    print(f"Finished {count} simultaneous submissions.")
    print(f"Success: {success}/{count}")
    if success == count:
        print("✅ PASSED Concurrency Check (No locks or 500s)")
    else:
        print("❌ FAILED Concurrency Check")

if __name__ == "__main__":
    run_adversarial()
    run_concurrency(15)
