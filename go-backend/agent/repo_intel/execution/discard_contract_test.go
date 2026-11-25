package execution

import("os";"path/filepath";"testing")

func TestDiscardSandboxContract(t *testing.T){
 root:=t.TempDir();svc,e:=NewMVPService(root,filepath.Join(root,"ledger.json"));if e!=nil{t.Fatal(e)}
 tmp,e:=EnsureTmpDir(root);if e!=nil{t.Fatal(e)}
 sandbox,e:=os.MkdirTemp(tmp,"mvp-sandbox-");if e!=nil{t.Fatal(e)}
 svc.executions["owned"]=&MVPExecutionResult{SandboxDir:sandbox,RollbackAvailable:true}
 ok,e:=svc.RollbackSandbox("owned");if e!=nil||!ok{t.Fatalf("discard: %v",e)}
 if _,e=os.Stat(sandbox);!os.IsNotExist(e){t.Fatal("sandbox remains")}
 if svc.executions["owned"].CleanupOperation!="DISCARD_SANDBOX"{t.Fatal("wrong contract")}
 if _,e=svc.RollbackSandbox("owned");e!=nil{t.Fatal("repeat discard should be idempotent")}
 outside:=t.TempDir();svc.executions["outside"]=&MVPExecutionResult{SandboxDir:outside,RollbackAvailable:true}
 if _,e=svc.RollbackSandbox("outside");e==nil{t.Fatal("unmanaged deletion accepted")}
 if _,e=os.Stat(outside);e!=nil{t.Fatal("unmanaged source touched")}
}
