"""Stage explicit mutations; only copies/reads source, never executes it."""
import difflib, hashlib, json, pathlib, shutil
ROOT=pathlib.Path('E:/SWE'); BASE=ROOT/'.tmp/reality-repos/mlsc-timetable'; DEST=ROOT/'.tmp/reality-trials'
resource='''package realityfixture
import("io";"os")
func Read(path string)([]byte,error){
 f,err:=os.Open(path)
 if err!=nil{return nil,err}
 defer f.Close()
 return io.ReadAll(f)
}
'''
sql='''package realityfixture
import("context";"database/sql")
func Find(ctx context.Context, db *sql.DB, name string)(*sql.Rows,error){
 query := "SELECT name FROM users WHERE name = ?"
 return db.QueryContext(ctx,query,name)
}
'''
error='''package realityfixture
import "encoding/json"
func Decode(data []byte) error {
 var value map[string]interface{}
 return json.Unmarshal(data,&value)
}
'''
command='''package realityfixture
import "os/exec"
func Run(value string) error {
 return exec.Command("echo",value).Run()
}
'''
resource_test='''package realityfixture
import("os";"path/filepath";"runtime";"testing")
func TestIndependentLeak(t *testing.T){
 if runtime.GOOS!="linux"{t.Skip("requires Linux /proc inside hardened container")}
 p:=filepath.Join(t.TempDir(),"input");if e:=os.WriteFile(p,[]byte("ok"),0600);e!=nil{t.Fatal(e)}
 before,e:=os.ReadDir("/proc/self/fd");if e!=nil{t.Fatal(e)}
 for i:=0;i<32;i++{b,e:=Read(p);if e!=nil||string(b)!="ok"{t.Fatalf("valid behavior: %s %v",b,e)}}
 after,e:=os.ReadDir("/proc/self/fd");if e!=nil{t.Fatal(e)}
 if len(after)>len(before)+2{t.Fatalf("descriptor leak: %d -> %d",len(before),len(after))}
}
'''
sql_test='''package realityfixture
import("context";"database/sql";"database/sql/driver";"errors";"io";"testing")
type probeDriver struct{}
type probeConn struct{}
type probeRows struct{}
var observedQuery string
var observedArgs []driver.NamedValue
func(probeDriver) Open(string)(driver.Conn,error){return probeConn{},nil}
func(probeConn) Prepare(string)(driver.Stmt,error){return nil,errors.New("unexpected prepare")}
func(probeConn) Close()error{return nil}
func(probeConn) Begin()(driver.Tx,error){return nil,errors.New("unexpected transaction")}
func(probeConn) QueryContext(_ context.Context,q string,args []driver.NamedValue)(driver.Rows,error){observedQuery=q;observedArgs=args;return probeRows{},nil}
func(probeRows) Columns()[]string{return []string{"name"}}
func(probeRows) Close()error{return nil}
func(probeRows) Next([]driver.Value)error{return io.EOF}
func TestIndependentSQLBinding(t *testing.T){
 sql.Register("reality-probe",probeDriver{});db,e:=sql.Open("reality-probe","");if e!=nil{t.Fatal(e)};defer db.Close()
 for _,value:=range []string{"Alice","' OR 1=1 --"}{
 rows,e:=Find(context.Background(),db,value);if e!=nil{t.Fatal(e)};rows.Close()
 if observedQuery!="SELECT name FROM users WHERE name = ?" || len(observedArgs)!=1 || observedArgs[0].Value!=value{t.Fatalf("unbound user input: query=%q args=%v",observedQuery,observedArgs)}
 }
}
'''
error_test='''package realityfixture
import "testing"
func TestIndependentErrorPropagation(t *testing.T){
 if Decode([]byte("invalid-json"))==nil{t.Fatal("malformed JSON silently accepted")}
 if e:=Decode([]byte(`{"valid":true}`));e!=nil{t.Fatal(e)}
}
'''
cases=[('RESOURCE_LEAK',resource,resource.replace(' defer f.Close()\n',''),resource_test),
 ('SQL_INJECTION',sql,sql.replace('"database/sql"','"database/sql";"fmt"').replace('"SELECT name FROM users WHERE name = ?"','fmt.Sprintf("SELECT name FROM users WHERE name = \'%s\'",name)').replace('db.QueryContext(ctx,query,name)','db.QueryContext(ctx,query)'),sql_test),
 ('UNHANDLED_ERROR',error,error.replace('return json.Unmarshal(data,&value)','json.Unmarshal(data,&value)\n return nil'),error_test),
 ('COMMAND_INJECTION',command,command.replace('exec.Command("echo",value)','exec.Command("sh","-c","echo "+value)'),None)]
manifest=[];DEST.mkdir(exist_ok=True)
for category,clean,mutated,reproducer in cases:
 target=DEST/category
 if target.exists(): raise RuntimeError('refuse overwrite '+str(target))
 shutil.copytree(BASE,target,ignore=shutil.ignore_patterns('.git'))
 fixture=target/'realityfixture';fixture.mkdir()
 (fixture/'go.mod').write_text('module realityfixture\n\ngo 1.25.0\n')
 (fixture/'defect.go').write_text(mutated)
 if reproducer:(fixture/'independent_test.go').write_text(reproducer)
 patch=''.join(difflib.unified_diff(clean.splitlines(True),mutated.splitlines(True),fromfile='clean/defect.go',tofile='mutated/defect.go'))
 evidence=ROOT/'reports/self_healing_reality_check/evidence'
 (evidence/(category+'-mutation.diff')).write_text(patch)
 (evidence/(category+'-baseline.go.txt')).write_text(clean)
 if reproducer:(evidence/(category+'-independent_test.go.txt')).write_text(reproducer)
 manifest.append({'category':category,'upstream':'Zylus08/mlsc-timetable','commit':'cbd082eafd24b290391145cb0399e81c6b41303c','kind':'added isolated controlled fixture package','original_source_sha256':hashlib.sha256(clean.encode()).hexdigest(),'mutated_source_sha256':hashlib.sha256(mutated.encode()).hexdigest(),'mutation_patch':category+'-mutation.diff','original_reproducer_status':'NOT_RUN_DOCKER_UNAVAILABLE','post_patch_status':'NOT_RUN','expected_action':'ABSTAIN' if category=='COMMAND_INJECTION' else 'REPAIR_OR_SUPPORTED_ABSTENTION'})
(ROOT/'reports/self_healing_reality_check/evidence/injection-manifest.json').write_text(json.dumps(manifest,indent=2))
