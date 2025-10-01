import re
import sublime
import sublime_plugin

charsets = [
  ('(', ')'),
  ('[', ']'),
  ('{', '}'),
  ('(', ')', ';'),
  ('(', ')', '|'),
  ('[', ']', ';'),
  ('[', ']', '|'),
  ('{', '}', ';'),
  ('{', '}', '|'),
  ('(', ')', '[', ']', ';'),
  ('(', ')', '[', ']', '|'),
  ('(', ')', '{', '}', ';'),
  ('(', ')', '{', '}', '|'),
  ('[', ']', '(', ')', ';'),
  ('[', ']', '(', ')', '|'),
  ('[', ']', '{', '}', ';'),
  ('[', ']', '{', '}', '|'),
  ('{', '}', '(', ')', ';'),
  ('{', '}', '(', ')', '|'),
  ('{', '}', '[', ']', ';'),
  ('{', '}', '[', ']', '|'),
]

charset_choices = [' '.join(charset) for charset in charsets]

class mox_swap_chars(sublime_plugin.TextCommand):
  def run(self, edit):
    view = self.view
    window = view.window()

    def on_select_from(index_from):
      if index_from == -1:
        return

      def on_select_to(index_to):
        if index_to == -1:
          return
        view.run_command('mox_swap_chars_from_to', {'index_from': index_from, 'index_to': index_to})

      select_charset(window, on_select_to, 'Select target charset; length must match source set')

    select_charset(window, on_select_from, 'Select source charset; must match current view')

class mox_swap_chars_from_to(sublime_plugin.TextCommand):
  def is_visible(self):
    return False

  def run(self, edit, index_from, index_to):
    view = self.view
    conf_from = charsets[index_from]
    conf_to = charsets[index_to]
    reg = re.compile('(?:{})'.format('|'.join('({})'.format(re.escape(val)) for val in conf_from)))

    for region in selections_or_buffer(view):
      def swap(match):
        return conf_to[match.lastindex - 1]

      result = reg.sub(swap, view.substr(region))
      view.replace(edit, region, result)

def select_charset(window, done, placeholder):
  window.show_quick_panel(
    items       = charset_choices,
    on_select   = done,
    placeholder = placeholder,
    flags       = sublime.MONOSPACE_FONT,
  )

def selections_or_buffer(view):
  sel = view.sel()
  if len(sel) == 0 or (len(sel) == 1 and sel[0].empty()):
    return [sublime.Region(0, view.size())]
  return sel
