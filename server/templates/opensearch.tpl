<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
	<ShortName>Hister</ShortName>
	<Description>Hister - web history on steroids</Description>
	<InputEncoding>UTF-8</InputEncoding>
	<Image type="image/x-icon">{{ .Config.BaseURL "/favicon.ico" }}</Image>
	<Image width="16" height="16">{{ .Config.BaseURL "/favicon.ico" }}</Image>
	<Query role="example" searchTerms="gpl"/>
	<Url type="text/html" template="{{ .Config.BaseURL "/" }}?q={searchTerms}" />
	<!-- <Url type="application/x-suggestions+json" rel="suggestions" template="{{ .Config.BaseURL "/suggestions" }}?q={searchTerms}"/> -->
</OpenSearchDescription>
