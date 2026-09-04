package report

import (
	"encoding/json"
	"html/template"
	"os"

	"github.com/artofdream/logify/internal/analyzer"
)

func Write(path string, r analyzer.Result) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return page.Execute(f, template.JS(b))
}

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width">
<title>Logify timeline</title>
<style>
:root{color-scheme:dark;--bg:#09101f;--panel:#151e32;--muted:#aebbd3;--text:#edf3ff;--line:#3b4b70;--accent:#7bb6ff;--success:#50d6a5}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px system-ui,sans-serif}
header{padding:24px 5vw;background:linear-gradient(120deg,#17274c,#10172a);border-bottom:1px solid var(--line)}
h1{margin:0 0 8px}
h2{margin:0}
.muted,.when,.src,.detail{color:var(--muted)}
main{max-width:1400px;margin:auto;padding:20px 5vw}
.stats,.section-heading,.issue-heading,.event-actions{display:flex;align-items:center;gap:10px;flex-wrap:wrap}
.stats{margin:15px 0}
.stat,.event,.issue,.notice,input,select,button{background:var(--panel);border:1px solid var(--line);border-radius:9px}
.stat{padding:10px 14px}
.notice{padding:12px 14px;margin:0 0 20px;border-left:4px solid var(--accent)}
.section-heading{justify-content:space-between;margin:24px 0 10px}
.controls{display:grid;grid-template-columns:2fr repeat(3,minmax(130px,1fr));gap:9px;margin-bottom:16px}
label{display:grid;gap:5px;font-weight:600}
input,select,button{padding:10px;color:var(--text);font:inherit}
button{cursor:pointer;font-weight:700}
button:hover:not(:disabled){border-color:var(--accent)}
button:disabled{cursor:default;opacity:.75}
button:focus-visible,input:focus-visible,select:focus-visible{outline:3px solid var(--accent);outline-offset:2px}
.event{display:grid;grid-template-columns:175px 80px 145px minmax(0,1fr);gap:12px;border-left:4px solid #718096;margin:8px 0;padding:12px;scroll-margin-top:16px}
.event.ERROR,.event.FATAL{border-left-color:#ff6077}
.event.WARN{border-left-color:#ffc24b}
.event.INFO{border-left-color:var(--success)}
.sev{font-weight:700}
.msg{white-space:pre-wrap;overflow-wrap:anywhere}
.detail{font-size:12px;margin-top:6px;overflow-wrap:anywhere}
.event-actions{margin-top:10px}
.issue{padding:14px;margin:10px 0;border-left:4px solid var(--success)}
.issue-heading{justify-content:space-between;margin-bottom:12px}
.issue-id{font:12px ui-monospace,SFMono-Regular,Consolas,monospace;color:var(--muted);overflow-wrap:anywhere}
.issue-title{max-width:720px}
.evidence{display:grid;grid-template-columns:max-content minmax(0,1fr);gap:6px 14px;margin:12px 0 0}
.evidence dt{font-weight:700;color:var(--muted)}
.evidence dd{margin:0;overflow-wrap:anywhere}
.feedback{min-height:20px;color:var(--success);font-weight:700}
.empty{text-align:center;padding:36px;color:var(--muted)}
@media(max-width:800px){.controls,.event{grid-template-columns:1fr}.evidence{grid-template-columns:1fr}.evidence dd{margin-bottom:6px}}
</style>
</head>
<body>
<header>
  <h1>Logify incident timeline</h1>
  <div id="meta" class="muted"></div>
</header>
<main>
  <div id="stats" class="stats" aria-label="Analysis summary"></div>
  <p class="notice" id="storage-notice"><strong>Issue storage:</strong> Issues exist only in this open report tab. They are not saved or transmitted. Export and import are not available yet; closing or reloading the page discards them after a browser warning.</p>

  <section aria-labelledby="issues-heading">
    <div class="section-heading">
      <h2 id="issues-heading">Issues (<span id="issue-count">0</span>)</h2>
    </div>
    <div id="issue-feedback" class="feedback" role="status" aria-live="polite"></div>
    <div id="issues"><div class="empty">No issues created. Use “Create issue” on timeline evidence.</div></div>
  </section>

  <section aria-labelledby="timeline-heading">
    <div class="section-heading"><h2 id="timeline-heading">Timeline</h2></div>
    <div class="controls">
      <label>Search<input id="q" type="search" placeholder="Messages, files, instances…"></label>
      <label>Severity<select id="sev"><option value="">All severities</option></select></label>
      <label>Instance<select id="inst"><option value="">All instances</option></select></label>
      <label>Source<select id="src"><option value="">All sources</option></select></label>
    </div>
    <div id="timeline"></div>
  </section>
</main>
<script>
const data={{.}};
const select=(query)=>document.querySelector(query);
const controls={q:select('#q'),sev:select('#sev'),inst:select('#inst'),src:select('#src')};
const issues=new Map();
let dirty=false;

function element(tag,className,text){
  const value=document.createElement(tag);
  if(className)value.className=className;
  if(text!==undefined)value.textContent=String(text);
  return value;
}

function replaceChildren(parent,children){
  parent.replaceChildren(...children);
}

function formatTime(value){
  return value?new Date(value).toLocaleString():'No timestamp';
}

function digestSuffix(evidenceID){
  return String(evidenceID).replace(/^evidence-v1-/, '');
}

function issueID(event){
  return 'issue-v1-'+digestSuffix(event.evidenceId);
}

function evidenceAnchor(event){
  return 'evidence-'+digestSuffix(event.evidenceId);
}

function defaultTitle(message){
  const first=String(message??'').split(/\r?\n/,1)[0].trim();
  return (first||'Follow up on timeline evidence').slice(0,200);
}

function addOptions(control,values){
  [...new Set(values)].sort().forEach((value)=>{
    const option=element('option','',value);
    option.value=value;
    control.append(option);
  });
}

function renderStats(){
  const values=[
    [data.events.length,'unique events'],
    [data.filesScanned,'files'],
    [data.warnings.length,'warnings'],
    [issues.size,'tracked issues']
  ];
  replaceChildren(select('#stats'),values.map(([count,label])=>{
    const item=element('div','stat');
    item.append(element('b','',count),document.createTextNode(' '+label));
    return item;
  }));
}

function showFeedback(message){
  select('#issue-feedback').textContent=message;
}

function createIssue(event){
  const id=issueID(event);
  if(issues.has(id)){
    showFeedback('Issue already exists; selected '+id+'.');
    const existing=select('#title-'+digestSuffix(event.evidenceId));
    if(existing)existing.focus();
    return;
  }
  issues.set(id,{id,title:defaultTitle(event.message),evidence:event});
  dirty=true;
  renderIssues();
  renderTimeline();
  renderStats();
  showFeedback('Created '+id+' from timeline evidence.');
  select('#title-'+digestSuffix(event.evidenceId)).focus();
}

function evidenceField(list,label,value){
  list.append(element('dt','',label),element('dd','',value));
}

function showEvidence(event){
  Object.values(controls).forEach((control)=>{control.value='';});
  renderTimeline();
  const target=document.getElementById(evidenceAnchor(event));
  if(target){
    target.focus();
    target.scrollIntoView({block:'center'});
  }
}

function renderIssues(){
  select('#issue-count').textContent=String(issues.size);
  if(issues.size===0){
    replaceChildren(select('#issues'),[element('div','empty','No issues created. Use “Create issue” on timeline evidence.')]);
    return;
  }
  const cards=[];
  issues.forEach((issue)=>{
    const event=issue.evidence;
    const card=element('article','issue');
    card.id='issue-'+digestSuffix(event.evidenceId);

    const heading=element('div','issue-heading');
    heading.append(element('span','issue-id',issue.id));
    const evidenceButton=element('button','','Show evidence');
    evidenceButton.type='button';
    evidenceButton.addEventListener('click',()=>showEvidence(event));
    heading.append(evidenceButton);
    card.append(heading);

    const label=element('label','issue-title','Issue title');
    const input=element('input');
    input.id='title-'+digestSuffix(event.evidenceId);
    input.type='text';
    input.maxLength=200;
    input.value=issue.title;
    input.addEventListener('input',()=>{
      issue.title=input.value;
      dirty=true;
      showFeedback('Updated title for '+issue.id+' in this report tab.');
    });
    label.append(input);
    card.append(label);

    const evidence=element('dl','evidence');
    evidenceField(evidence,'Evidence ID',event.evidenceId);
    evidenceField(evidence,'Signature',event.signature);
    evidenceField(evidence,'Instance',event.instance);
    evidenceField(evidence,'Source',event.file+':'+event.line);
    evidenceField(evidence,'First seen',formatTime(event.firstSeen));
    evidenceField(evidence,'Last seen',formatTime(event.lastSeen));
    evidenceField(evidence,'Occurrences',event.occurrences);
    card.append(evidence);
    cards.push(card);
  });
  replaceChildren(select('#issues'),cards);
}

function severityClass(value){
  return ['TRACE','DEBUG','INFO','WARN','ERROR','FATAL','UNKNOWN'].includes(value)?value:'UNKNOWN';
}

function renderTimeline(){
  const query=controls.q.value.toLowerCase();
  const filtered=data.events.filter((event)=>
    (!query||[event.message,event.file,event.instance].join(' ').toLowerCase().includes(query))&&
    (!controls.sev.value||event.severity===controls.sev.value)&&
    (!controls.inst.value||event.instance===controls.inst.value)&&
    (!controls.src.value||event.sourceType===controls.src.value)
  );
  if(filtered.length===0){
    replaceChildren(select('#timeline'),[element('div','empty','No matching events.')]);
    return;
  }
  const events=filtered.map((event)=>{
    const article=element('article','event '+severityClass(event.severity));
    article.id=evidenceAnchor(event);
    article.tabIndex=-1;
    article.append(
      element('div','when',event.hasTimestamp?formatTime(event.timestamp):'No timestamp'),
      element('div','sev',event.severity),
      element('div','src',event.instance+'\n'+event.sourceType)
    );

    const body=element('div');
    body.append(element('div','msg',event.message));
    let details=event.file+':'+event.line+' • '+event.signature;
    if(event.occurrences>1){
      details+=' • '+event.occurrences+' occurrences; first '+formatTime(event.firstSeen)+'; last '+formatTime(event.lastSeen);
    }
    body.append(element('div','detail',details));

    const actions=element('div','event-actions');
    const button=element('button','',issues.has(issueID(event))?'Issue created':'Create issue');
    button.type='button';
    button.disabled=issues.has(issueID(event));
    button.setAttribute('aria-label','Create issue from '+event.file+' line '+event.line);
    button.addEventListener('click',()=>createIssue(event));
    actions.append(button);
    body.append(actions);
    article.append(body);
    return article;
  });
  replaceChildren(select('#timeline'),events);
}

addOptions(controls.sev,data.events.map((event)=>event.severity));
addOptions(controls.inst,data.events.map((event)=>event.instance));
addOptions(controls.src,data.events.map((event)=>event.sourceType));
select('#meta').textContent=data.root+' • generated '+new Date(data.generatedAt).toLocaleString();
Object.values(controls).forEach((control)=>control.addEventListener('input',renderTimeline));
window.addEventListener('beforeunload',(event)=>{
  if(!dirty||issues.size===0)return;
  event.preventDefault();
  event.returnValue='';
});
renderStats();
renderIssues();
renderTimeline();
</script>
</body>
</html>`))
