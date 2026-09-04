package report

import (
	"encoding/json"
	"github.com/artofdream/logify/internal/analyzer"
	"html/template"
	"os"
)

func Write(path string, r analyzer.Result) error {
	if r.Events == nil {
		r.Events = []analyzer.Event{}
	}
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	b, e := json.Marshal(r)
	if e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return page.Execute(f, template.JS(b))
}

var page = template.Must(template.New("p").Parse(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Logify timeline</title><style>:root{color-scheme:dark;--bg:#09101f;--panel:#151e32;--muted:#97a7c4;--text:#edf3ff;--line:#2b3858}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px system-ui}header{padding:24px 5vw;background:linear-gradient(120deg,#17274c,#10172a);border-bottom:1px solid var(--line)}h1{margin:0 0 8px}.muted,.when,.src,.detail{color:var(--muted)}main{max-width:1400px;margin:auto;padding:20px 5vw}.stats{display:flex;gap:10px;margin:15px 0;flex-wrap:wrap}.stat,.event,input,select{background:var(--panel);border:1px solid var(--line);border-radius:9px}.stat{padding:10px 14px}.controls{display:grid;grid-template-columns:2fr repeat(3,1fr);gap:9px;margin-bottom:16px}input,select{padding:10px;color:var(--text)}.event{display:grid;grid-template-columns:175px 80px 145px 1fr;gap:12px;border-left:4px solid #718096;margin:8px 0;padding:12px}.event.ERROR,.event.FATAL{border-left-color:#ff6077}.event.WARN{border-left-color:#ffc24b}.event.INFO{border-left-color:#50d6a5}.sev{font-weight:700}.msg{white-space:pre-wrap;overflow-wrap:anywhere}.detail{font-size:12px;margin-top:6px}.empty{text-align:center;padding:50px;color:var(--muted)}@media(max-width:800px){.controls,.event{grid-template-columns:1fr}}</style></head><body><header><h1>Logify incident timeline</h1><div id="meta" class="muted"></div></header><main><div id="stats" class="stats"></div><div class="controls"><input id="q" placeholder="Search messages, files, instances…"><select id="sev"><option value="">All severities</option></select><select id="inst"><option value="">All instances</option></select><select id="src"><option value="">All sources</option></select></div><div id="timeline"></div></main><script>const d={{.}},$=x=>document.querySelector(x),e=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const els={q:$('#q'),sev:$('#sev'),inst:$('#inst'),src:$('#src')};function opts(x,a){[...new Set(a)].sort().forEach(v=>x.insertAdjacentHTML('beforeend','<option>'+e(v)+'</option>'))}opts(els.sev,(d.events||[]).map(x=>x.severity));opts(els.inst,(d.events||[]).map(x=>x.instance));opts(els.src,(d.events||[]).map(x=>x.sourceType));$('#meta').textContent=d.root+' • generated '+new Date(d.generatedAt).toLocaleString();$('#stats').innerHTML='<div class="stat"><b>'+(d.events||[]).length+'</b> unique events</div><div class="stat"><b>'+d.filesScanned+'</b> files</div><div class="stat"><b>'+(d.warnings||[]).length+'</b> warnings</div>';function render(){let q=els.q.value.toLowerCase(),a=(d.events||[]).filter(x=>(!q||[x.message,x.file,x.instance].join(' ').toLowerCase().includes(q))&&(!els.sev.value||x.severity===els.sev.value)&&(!els.inst.value||x.instance===els.inst.value)&&(!els.src.value||x.sourceType===els.src.value));$('#timeline').innerHTML=a.length?a.map(x=>'<article class="event '+e(x.severity)+'"><div class="when">'+(x.hasTimestamp?new Date(x.timestamp).toLocaleString():'No timestamp')+'</div><div class="sev">'+e(x.severity)+'</div><div class="src">'+e(x.instance)+'<br>'+e(x.sourceType)+'</div><div><div class="msg">'+e(x.message)+'</div><div class="detail">'+e(x.file)+':'+x.line+' • '+e(x.signature)+(x.occurrences>1?' • '+x.occurrences+' occurrences; last '+new Date(x.lastSeen).toLocaleString():'')+'</div></div></article>').join(''):'<div class="empty">No matching events.</div>'}Object.values(els).forEach(x=>x.addEventListener('input',render));render()</script></body></html>`))
