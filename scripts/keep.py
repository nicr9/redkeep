"""
Script that converts Google Keep notes from a Google Takeout dump.

Photograph notes are currently ignored.

Usage: python3 scripts/keep.py Takeout/Keep
"""
from sys import argv
from os import walk
from os.path import join as path_join
from collections import defaultdict
from pprint import pprint
from time import strptime, mktime

from bs4 import BeautifulSoup
from bs4.element import NavigableString, Tag

import yaml

# Workaround to represent body as literal scalar blocks in YAML
class literal(str): pass
def literal_repr(dumper, data):
        return dumper.represent_scalar(u'tag:yaml.org,2002:str', data, style='|')
yaml.add_representer(literal, literal_repr)

def get_title(soup):
    title = soup.body.find(**{'class': 'title'})
    return "" if title is None else title.text

def parse_note(soup):
    content = soup.body.find(**{'class': 'content'})
    note = parse_contents(content)
    note['body'] = literal(note['body'])
    note['created'] = note['updated'] = get_timestamp(soup)
    note['title'] = get_title(soup)
    note['tags'] = get_labels(soup)
    note['id'] = ""

    return dict(note)

def parse_contents(tag):
    results = defaultdict(list)
    for elem in tag.contents:
        if type(elem) == NavigableString:
            results['body'].append(str(elem))
        elif type(elem) == Tag:
            if elem.name == 'br':
                results['body'].append('\n')
            elif elem.name == 'div':
                status = 'open' if elem.div.text == "\u2610" else 'closed'
                sub = elem.find(**{'class': 'text'})
                parsed = parse_contents(sub)
                body = parsed['body']
                results[status].append(body)
                if len(parsed.keys()) > 1:
                    import pdb; pdb.set_trace()

    results['body'] = ''.join(results['body']).strip()
    return results

def get_labels(soup):
    labels = [label.text for label in soup.body.find_all(**{'class': 'label'})]
    if soup.body.find(**{'class': 'archived'}):
        labels.append('archived')

    return labels

def get_timestamp(soup):
    heading = soup.body.find(**{'class': 'heading'}).text
    timestamp = strptime(heading, "\n%d %b %Y, %H:%M:%S")
    return str(int(mktime(timestamp)))

if __name__ == "__main__":
    keep = argv[1]

    for root, _, notes in walk(keep):
        photos = []
        for note in notes:
            relpath = path_join(root, note)
            with open(relpath, 'r') as inp, open('keep.yaml', 'a') as outp:
                try:
                    soup = BeautifulSoup(inp.read(), 'html')
                except UnicodeDecodeError:
                    photos.append(relpath)
                    continue
                print(relpath)
                contents = parse_note(soup)
                yaml.dump([contents], outp)
