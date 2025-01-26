all

# Extend line length, since each sentence should be on a separate line.
rule 'MD013', :line_length => 99999

# https://github.com/updownpress/markdown-lint/blob/master/rules/024-no-duplicate-header.md
exclude_rule 'MD024'

# https://github.com/updownpress/markdown-lint/blob/master/rules/033-no-inline-html.md
exclude_rule 'MD033'
