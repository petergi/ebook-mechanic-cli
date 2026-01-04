from pathlib import Path
import re
import sys

root = Path('.')
md_files = list(root.rglob('*.md'))
link_re = re.compile(r'\[[^\]]*\]\(([^)]+)\)')

missing = []
for md in md_files:
    text = md.read_text(encoding='utf-8')
    for link in link_re.findall(text):
        if link.startswith(('http://', 'https://', 'mailto:')):
            continue
        if link.startswith('#'):
            continue
        link_path = link.split('#', 1)[0].strip()
        if not link_path:
            continue
        if link_path.startswith('data:'):
            continue
        target = (md.parent / link_path).resolve()
        if not target.exists():
            missing.append((md, link))

if missing:
    print('Missing links:')
    for md, link in missing:
        print(f'- {md}: {link}')
    sys.exit(1)

print('All local markdown links resolve.')
