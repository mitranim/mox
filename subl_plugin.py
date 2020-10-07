import re
import sublime
import sublime_plugin

pairs = {
    '(': '{',
    ')': '}',
    '{': '(',
    '}': ')',
}

DELIM_RE = re.compile('[{}]'.format(''.join('\\' + char for char in pairs.keys())))

def swap_delims(match):
    return pairs[match.group(0)]

def selections_or_buffer(view):
    sel = view.sel()
    if len(sel) == 0 or (len(sel) == 1 and sel[0].empty()):
        return [sublime.Region(0, view.size())]
    return sel

def replace_in_region(view, edit, region, repl):
    source = view.substr(region)
    result = DELIM_RE.sub(string=source, repl=repl)
    view.replace(edit, region, result)

class mox_swap_parens_braces(sublime_plugin.TextCommand):
    # def is_enabled(self):
    #     view = self.view
    #     return view.score_selector(0, 'source.mox') > 0

    def run(self, edit):
        view = self.view
        for region in selections_or_buffer(view):
            replace_in_region(view, edit, region, swap_delims)
