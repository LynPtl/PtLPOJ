import os
import glob
import json
import ast

# 设定绝对路径
BASE_DIR = "/home/pt/PtLPOJ"
SOURCE_DIR = os.path.join(BASE_DIR, "sample_questions")
OUT_DIR = os.path.join(BASE_DIR, "data", "problems")

def extract_content(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. 直接截断掉 #sample answer 往后的所有内容 (安全防泄漏边界)
    if '#sample answer' in content:
        content = content.split('#sample answer')[0]

    tree = ast.parse(content)
    
    imports = []
    func_node = None
    
    # 2. 遍历 AST，找到 Import 和主要的函数 (题目通常叫 f)
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
    
    # 3. 提取函数的签名 (Signature，比如 def f(height):)
    lines = content.split('\n')
    sig_start = func_node.lineno - 1
    sig_lines = []
    for i in range(sig_start, len(lines)):
        sig_lines.append(lines[i])
        if ':' in lines[i]:
            break
    signature = '\n'.join(sig_lines).split(':')[0] + ':'

    # 4. 解析 doctest 中的所有例子，智能隔离公开/隐藏用例
    import doctest
    parser = doctest.DocTestParser()
    try:
        parsed = parser.parse(docstring)
    except Exception:
        parsed = [docstring]
    
    examples = [item for item in parsed if isinstance(item, doctest.Example)]
    
    # 动态计算公开用例数量：至少2个，或者占据总用例的约 60%
    # 这样既能保证规律可以被看出，又能有效隐藏后半部分(边界/大压力测试点)
    total_cases = len(examples)
    public_count = total_cases if total_cases <= 3 else int(total_cases * 0.6)
    # 取一个保底值，只要大于2个用例，我们至少展示3个以上帮助理解
    public_count = max(3, public_count) if total_cases > 3 else total_cases
    
    public_docstring_parts = []
    example_idx = 0
    for item in parsed:
        if isinstance(item, str):
            public_docstring_parts.append(item)
        elif isinstance(item, doctest.Example):
            # 取出前 N 个作为公开用例 (Public Cases)
            if example_idx < public_count:
                public_docstring_parts.append(f">>> {item.source}{item.want}")
            example_idx += 1
            
    public_docstring = "".join(public_docstring_parts).strip()
    
    # 5. 组合 Scaffold (代码脚手架)
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
        "full_docstring": docstring, # 完整测试用例交给沙盒
        "case_count": len(examples)
    }

def main():
    os.makedirs(OUT_DIR, exist_ok=True)
    
    problems_json = []
    files = sorted(glob.glob(os.path.join(SOURCE_DIR, "*.py")))
    
    print(f"Found {len(files)} python files in {SOURCE_DIR}")
    
    for i, file_path in enumerate(files):
        filename = os.path.basename(file_path)
        title = filename.replace('.py', '')
        
        prob_id = 1001 + i
        res = extract_content(file_path)
        if not res:
            print(f"  [Skip] No function found in {filename}")
            continue
            
        print(f" [✓] Processed: {prob_id} - {title} (Found {res['case_count']} test cases)")
            
        prob_dir = os.path.join(OUT_DIR, str(prob_id))
        os.makedirs(prob_dir, exist_ok=True)
        
        # 写入返回给前端看的脚手架
        with open(os.path.join(prob_dir, "scaffold.py"), 'w', encoding='utf-8') as f:
            f.write(res['scaffold'])
            
        # 写入服务器留存的全量测试用例（由于是隐式注入，保留完整 docstring 即可）
        with open(os.path.join(prob_dir, "tests.txt"), 'w', encoding='utf-8') as f:
            f.write(res['full_docstring'])
            
        # 写入题目描述 Markdown
        with open(os.path.join(prob_dir, "problem.md"), 'w', encoding='utf-8') as f:
            f.write(f"# {title}\n\n## 题目描述\n\n请实现函数 `{res['func_name']}`。\n\n## 示例框架\n\n```python\n{res['scaffold']}\n```\n")
            
        # 记录 Metadata 到全局题库
        problems_json.append({
            "id": prob_id,
            "title": title,
            "difficulty": "Easy",
            "tags": ["Auto-generated"],
            "time_limit_ms": 1000,
            "memory_limit_kb": 65536,
            "case_count": res['case_count']
        })
        
    with open(os.path.join(OUT_DIR, "problems.json"), 'w', encoding='utf-8') as f:
        json.dump(problems_json, f, ensure_ascii=False, indent=4)
        print("\n=> Successfully generated problems.json!")

if __name__ == '__main__':
    main()
