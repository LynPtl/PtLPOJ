import urllib.request
def f(x):
    urllib.request.urlopen('http://example.com', timeout=1)
    return x
