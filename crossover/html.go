package crossover

const inlineTemplate = `
<!DOCTYPE html>
<html>
<head>
<style>

pre {
    background: #f4f4f4;
    border: 1px solid #ddd;
    color: #666;
    page-break-inside: avoid;
    font: normal normal 14px/16px "Courier New",Courier,Monospace;
    line-height: 1.0;
    margin-bottom: 1.6em;
    max-width: 100%;
    padding: .5em 1em;
    display: block;
    word-wrap: break-word;
    margin-left: 30px;
    overflow: auto;
    overflow-x: auto;
    white-space: pre-wrap;
    counter-reset: line;
}

body {
    margin:20;
    font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;
    font-size:14px;
    line-height:1.4;
    color:#222;
}

container {
    padding:10px;
}

a {
    color:#444;
}

h1,h2,h3 {
    font-size:22px;
    font-weight:500;
    margin-top:0px;
    margin-bottom:10px;
    color:#444;
}

</style>
</head>
<body>
	{{range .}}<div class="container"><h3><a href="{{ .Link }}">{{ .Title }}</a></h3>{{ .Content }}</div>{{ end }}
</body>
</html>
`
