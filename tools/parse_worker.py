import sys
import json
import ast
import doctest

def extract_content(content):
    if '#sample answer' in content:
        content = content.split('#sample answer')[0]

    tree = ast.parse(content)
    
    imports = []
    func_node = None
    
    for node in tree.body:
        if isinstance(node, (ast.Import, ast.ImportFrom)):
            imports.append(ast.get_source_segment(content, node))
        elif isinstance(node, ast.FunctionDef):
            if func_node is None or node.name == 'f':
                func_node = node

    if not func_node:
        return None
        
    func_name = func_node.name
    docstring = ast.get_docstring(func_node) or ""
    
    lines = content.split('\n')
    sig_start = func_node.lineno - 1
    sig_lines = []
    for i in range(sig_start, len(lines)):
        sig_lines.append(lines[i])
        if ':' in lines[i]:
            break
    signature = '\n'.join(sig_lines).split(':')[0] + ':'

    parser = doctest.DocTestParser()
    try:
        parsed = parser.parse(docstring)
    except Exception:
        parsed = [docstring]
    
    examples = [item for item in parsed if isinstance(item, doctest.Example)]
    
    total_cases = len(examples)
    public_count = total_cases if total_cases <= 3 else int(total_cases * 0.6)
    public_count = max(3, public_count) if total_cases > 3 else total_cases
    
    public_docstring_parts = []
    example_idx = 0
    for item in parsed:
        if isinstance(item, str):
            public_docstring_parts.append(item)
        elif isinstance(item, doctest.Example):
            if example_idx < public_count:
                public_docstring_parts.append(f">>> {item.source}{item.want}")
            example_idx += 1
            
    public_docstring = "".join(public_docstring_parts).strip()
    
    scaffold = "\n".join(imports) + ("\n\n" if imports else "")
    scaffold += f"{signature}\n"
    scaffold += f"    \"\"\"\n"
    for line in public_docstring.split('\n'):
        if line.strip():
            scaffold += f"    {line}\n"
        else:
            scaffold += "\n"
    scaffold += f"    \"\"\"\n"
    scaffold += f"    # 请在此处输入你的代码...\n"
    scaffold += f"    pass\n"
    
    return {
        "func_name": func_name,
        "scaffold": scaffold,
        "full_docstring": docstring,
        "case_count": len(examples)
    }

if __name__ == "__main__":
    content = sys.stdin.read()
    res = extract_content(content)
    if res:
        print(json.dumps(res, ensure_ascii=False))
    else:
        print(json.dumps({"error": "No valid function f found"}), file=sys.stderr)
        sys.exit(1)
