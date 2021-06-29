// This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
// Copyright (c) 2021 John Dudmesh.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.

// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"fmt"
	"justthetalk/model"
	"regexp"
	"strings"
)

type PostFormat struct {
	re       *regexp.Regexp
	formatFn func(string) string
}

type PostFormatter struct {
	formatters      []PostFormat
	linkReplacer    *regexp.Regexp
	postNumReplacer *regexp.Regexp
	lineSplitter    *regexp.Regexp
}

func NewPostFormatter() *PostFormatter {

	formatter := &PostFormatter{
		formatters: []PostFormat{
			{
				re:       regexp.MustCompile("^&gt; "),
				formatFn: formatQuoted,
			},
			{
				re:       regexp.MustCompile("^s "),
				formatFn: formatStrikethrough,
			},
			{
				re:       regexp.MustCompile("^b "),
				formatFn: formatBold,
			},
			{
				re:       regexp.MustCompile("^i "),
				formatFn: formatItalic,
			},
			{
				re:       regexp.MustCompile("^u "),
				formatFn: formatUnderline,
			},
			{
				re:       regexp.MustCompile("^c "),
				formatFn: formatCentre,
			},
			{
				re:       regexp.MustCompile("^`"),
				formatFn: formatCode,
			},
			{
				re:       regexp.MustCompile("^\\*"),
				formatFn: formatBullet,
			},
			{
				re:       regexp.MustCompile("^]+ "),
				formatFn: formatIndent,
			},
			{
				re:       regexp.MustCompile("^\\} "),
				formatFn: formatLinebreak,
			},
			{
				re:       regexp.MustCompile("^\\| "),
				formatFn: formatSpoiler,
			},
		},
		linkReplacer:    regexp.MustCompile(`https?:\/\/([-\w\.]+)+(:\d+)?\S+\/?`),
		postNumReplacer: regexp.MustCompile(`&?(amp;)?#(\d+)`),
		lineSplitter:    regexp.MustCompile(`[\r\n]`),
	}

	return formatter
}

func (p *PostFormatter) ApplyPostFormatting(rawText string, discussion *model.Discussion) string {

	markup := "<div>"

	text := strings.TrimSpace(rawText)
	text = p.formatLinks(text)
	text = p.formatPostLinks(text, discussion)

	lines := strings.SplitAfter(text, "\n")
	for _, line := range lines {

		if len(strings.TrimSpace(line)) == 0 {
			markup += "<br/><br/>"
		} else {

			found := false
			for _, formatter := range p.formatters {
				if formatter.re.MatchString(line) {
					found = true
					markup += formatter.formatFn(line)
					break
				}
			}

			if !found {
				markup += line
			}

		}

	}

	markup += "</div>"

	return markup

}

func formatQuoted(text string) string {
	return fmt.Sprintf("<p class='post-quoted'>%s</p>", text[5:])
}

func formatStrikethrough(text string) string {
	return fmt.Sprintf("<span class='post-strikethrough'>%s</span>", text[2:])
}

func formatBold(text string) string {
	return fmt.Sprintf("<span class='post-bold'>%s</span>", text[2:])
}

func formatItalic(text string) string {
	return fmt.Sprintf("<span class='post-italic'>%s</span>", text[2:])
}

func formatUnderline(text string) string {
	return fmt.Sprintf("<span class='post-underline'>%s</span>", text[2:])
}

func formatCentre(text string) string {
	return fmt.Sprintf("<p class='post-centre'>%s</p>", text[2:])
}

func formatCode(text string) string {
	return fmt.Sprintf("<div style='margin-left: 40px;margin-right: 40px;'><code>%s</code></div>", text[2:])
}

func formatBullet(text string) string {
	return fmt.Sprintf("<ul class='post-noindentbullet'><li>%s</li></ul>", text[2:])
}

func formatIndent(text string) string {
	indentCount := strings.Index(text, " ")
	return fmt.Sprintf("<p style='margin-left: %dpx;'>%s</p>", indentCount*10, text[indentCount+1:])
}

func formatLinebreak(text string) string {
	return fmt.Sprintf("<p>%s</p>", text[2:])
}

func formatSpoiler(text string) string {
	return fmt.Sprintf("<div class='post-spoiler'><div class='post-spoiler-heading' onclick='this.parentElement.classList.add(\"show\");'>Spoiler (click to reveal)</div><div class='post-spoiler-body'>%s</div></div>", text[2:])
}

func (p *PostFormatter) formatLinks(text string) string {
	return p.linkReplacer.ReplaceAllString(text, "<a href=\"$0\" rel='nofollow'>$0</a>")
}

func (p *PostFormatter) formatPostLinks(text string, discussion *model.Discussion) string {

	match := p.postNumReplacer.FindStringSubmatchIndex(text)
	if match != nil {
		if text[match[0]] != '&' {
			nextText := fmt.Sprintf("%s<a href=\"%s/%s\">%s</a>", text[:match[0]], discussion.Url, text[match[4]:match[5]], text[match[0]:match[1]])
			return nextText + p.formatPostLinks(text[match[1]:], discussion)
		} else {
			return text[:match[1]] + p.formatPostLinks(text[match[1]:], discussion)
		}
	} else {
		return text
	}

}
