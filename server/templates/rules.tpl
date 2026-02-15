{{define "main"}}
{{ $CSRF := .CSRF }}
<div class="container full-width">
    <form method="post">
        <h2>Skip Rules</h2>
        <p>Define regexps to forbid indexing matching URLs</p>
        <textarea placeholder="Text..." name="skip" class="full-width" >{{ Join .Config.Rules.Skip.ReStrs "\n" }}</textarea>
        <br />
        <input type="submit" value="Save" class="mt-1" />
        <h2>Priority Rules</h2>
        <p>Define regexps to prioritize matching URLs</p>
        <textarea placeholder="Text..." name="priority" class="full-width" >{{ Join .Config.Rules.Priority.ReStrs "\n" }}</textarea>
        <br />
        <input type="hidden" value="{{ $CSRF }}" name="csrf_token" />
        <input type="submit" value="Save" class="mt-1" />
    </form>
        <h2>Search Keyword Aliases</h2>
        <p>Define aliases to simplify queries. Alias strings in queries are automatically replaced with the provided value.</p>
        {{ if .Config.Rules.Aliases }}
        <table class="mv-1">
            <tr><th>Keyword</th><th>Value</th><th>Delete</th></tr>
            {{ range $k, $v := .Config.Rules.Aliases }}
            <tr>
                <td>{{ $k }}</td>
                <td>{{ $v }}</td>
                <td>
                    <form action="/delete_alias" method="post">
                        <input type="hidden" value="{{ $k }}" name="alias" />
                        <input type="hidden" value="{{ $CSRF }}" name="csrf_token" />
                        <input type="submit" value="Delete" />
                    </form>
                </td>
            </tr>
            {{ end }}
        </table>
        {{ else }}
        <h3>There are no aliases</h3>
        {{ end }}
        <details>
            <summary class="mv-1">Add new alias</summary>
            <form action="add_alias" method="post">
                <input type="text" name="alias-keyword" placeholder="Keyword..."  class="full-width" />
                <input type="text" name="alias-value" placeholder="Value..."  class="full-width" />
                <br />
                <input type="hidden" value="{{ $CSRF }}" name="csrf_token" />
                <input type="submit" value="Save" class="mt-1" />
            </form>
        </details>
    </form>
</div>
{{end}}
