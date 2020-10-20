import re
import sublime
import sublime_plugin

delim_pairs = {
    '(': '{',
    ')': '}',
    '{': '(',
    '}': ')',
}

RE_COMMENT_DELIMS = re.compile('(?:{})'.format('|'.join(map(re.escape, delim_pairs.keys()))))

comment_delim_pairs = {
    '{{': '[',
    '}}': ']',
    '[': '{{',
    ']': '}}',
}

RE_COMMENT_DELIMS = re.compile('(?:{})'.format('|'.join(map(re.escape, comment_delim_pairs.keys()))))

def swap_delims(match):
    return delim_pairs[match.group(0)]

def swap_comment_delims(match):
    return comment_delim_pairs[match.group(0)]

def selections_or_buffer(view):
    sel = view.sel()
    if len(sel) == 0 or (len(sel) == 1 and sel[0].empty()):
        return [sublime.Region(0, view.size())]
    return sel

def replace_in_region(view, edit, region, reg, repl):
    source = view.substr(region)
    result = reg.sub(string=source, repl=repl)
    view.replace(edit, region, result)

class mox_swap_delims(sublime_plugin.TextCommand):
    def run(self, edit):
        view = self.view
        for region in selections_or_buffer(view):
            replace_in_region(view, edit, region, RE_DELIMS, swap_delims)

class mox_swap_comment_delims(sublime_plugin.TextCommand):
    def run(self, edit):
        view = self.view
        for region in selections_or_buffer(view):
            replace_in_region(view, edit, region, RE_COMMENT_DELIMS, swap_comment_delims)
