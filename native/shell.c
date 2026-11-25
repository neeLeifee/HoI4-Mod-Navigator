/* Native x64 COM handler. No CRT, no Go runtime inside Explorer. */
#include <stdint.h>
#include <stddef.h>
typedef uint16_t W;
typedef int32_t HR;
typedef uint32_t U;
typedef void *H;
typedef struct { U a; uint16_t b,c; uint8_t d[8]; } GUID;
#define API __declspec(dllimport)
#define OK ((HR)0)
#define NO ((HR)1)
#define FAIL ((HR)0x80004005)
#define INVALID ((HR)0x80070057)
#define OOM ((HR)0x8007000e)
#define NOIF ((HR)0x80004002)
#define NOTIMPL ((HR)0x80004001)
#define CLASSNA ((HR)0x80040111)
#define NOAGG ((HR)0x80040110)
#define HKCU ((H)(intptr_t)(int32_t)0x80000001)
#define MAXPATH 32768u
#define MAXCOMMANDS 80u
#define MAXSNAP (1024u*1024u)
#define WSTR(s) ((const W*)L##s)
API H GetProcessHeap(void);
API H HeapAlloc(H,U,size_t);
API int HeapFree(H,U,H);
API U GetFileAttributesW(const W*);
API int CompareStringOrdinal(const W*,int,const W*,int,int);
API int32_t RegGetValueW(H,const W*,const W*,U,U*,void*,U*);
API int SHGetPathFromIDListEx(const void*,W*,U,U);
API U DragQueryFileW(H,U,W*,U);
API void ReleaseStgMedium(void*);
API int InsertMenuW(H,U,U,uintptr_t,const W*);
API H ShellExecuteW(H,const W*,const W*,const W*,const W*,int);
static void *alloc(size_t n){return HeapAlloc(GetProcessHeap(),8,n);}
static void drop(void *p){if(p) HeapFree(GetProcessHeap(),0,p);}
static long live=0, locks=0; /* on Windows long is 32 bits */
static U inc(long *p){return (U)__atomic_add_fetch(p,1,__ATOMIC_SEQ_CST);}
static U dec(long *p){return (U)__atomic_sub_fetch(p,1,__ATOMIC_SEQ_CST);}
static int eqg(const GUID *a,const GUID*b){if(!a||!b)return 0;const uint8_t*x=(const uint8_t*)a,*y=(const uint8_t*)b;for(U i=0;i<16;i++)if(x[i]!=y[i])return 0;return 1;}
static const GUID iidUnknown={0,0,0,{0xc0,0,0,0,0,0,0,0x46}}, iidFactory={1,0,0,{0xc0,0,0,0,0,0,0,0x46}}, iidInit={0x000214e8,0,0,{0xc0,0,0,0,0,0,0,0x46}}, iidMenu={0x000214e4,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const GUID classes[3]={{0x76ae4d21,0xcad1,0x4b8e,{0xa7,0x9e,0x85,0x55,0xca,0xb5,0x71,0x01}},{0x76ae4d21,0xcad1,0x4b8e,{0xa7,0x9e,0x85,0x55,0xca,0xb5,0x71,0x02}},{0x76ae4d21,0xcad1,0x4b8e,{0xa7,0x9e,0x85,0x55,0xca,0xb5,0x71,0x03}}};
static U len(const W*s){U n=0;while(s[n])n++;return n;}
static int same(const W*a,const W*b){return CompareStringOrdinal(a,-1,b,-1,1)==2;}
static int within(const W*root,const W*path){U r=len(root),p=len(path);return r>=3 && p>=r && (p==r||path[r]=='\\') && CompareStringOrdinal(root,(int)r,path,(int)r,1)==2;}
/* Direct child of Documents/.../mod must be a real directory, not a .mod file. */
static int localwithin(const W*root,W*path,U scope){
 U r=len(root), n=len(path);if(!within(root,path)||n<=r+1)return 0;
 U end=r+1;while(end<n&&path[end]!='\\')end++;
 if(scope==2&&end==n)return 0;
 W saved=path[end];path[end]=0;U attr=GetFileAttributesW(path);path[end]=saved;
 return attr!=0xffffffffu&&(attr&0x10u)&&!(attr&0x400u);
}
static int token(const W *s){U n=len(s);if(!n||n>40)return 0;for(U i=0;i<n;i++)if(!((s[i]>='a'&&s[i]<='z')||(s[i]>='0'&&s[i]<='9')||s[i]=='-'))return 0;return 1;}
static int pathok(const W*s){U n=len(s);if(n<3||s[1]!=':'||s[2]!='\\')return 0;for(U i=0;i<n;i++)if(s[i]=='"'||s[i]=='\r'||s[i]=='\n')return 0;return 1;}
static W* line(W**next,W*end){W*p=*next;if(p>=end)return 0;W*q=p;while(q<end&&*q!='\n'&&*q)q++;if(q>=end||*q!='\n')return 0;*q=0;*next=q+1;return p;}
typedef struct {const W*label,*action,*id;} Entry;
typedef struct Obj Obj;
typedef struct {const void*vt;Obj*owner;} Iface;
struct Obj{Iface init,menu;long ref;U scope,count,available;W*data;const W*exe;Entry entries[MAXCOMMANDS],shown[MAXCOMMANDS];W path[MAXPATH];};
static HR qi(Iface*self,const GUID*i,void**out){if(!out)return INVALID;*out=0;Obj*o=self->owner;if(eqg(i,&iidUnknown)||eqg(i,&iidInit))*out=&o->init;else if(eqg(i,&iidMenu))*out=&o->menu;else return NOIF;inc(&o->ref);return OK;}
static U add(Iface*self){return inc(&self->owner->ref);}
static U release(Iface*self){Obj*o=self->owner;U n=dec(&o->ref);if(!n){drop(o->data);drop(o);dec(&live);}return n;}
/* FORMATETC and STGMEDIUM: exact Win64 ABI. */
typedef struct {uint16_t format;void*ptd;U aspect;int32_t index;U tymed;} Format;
typedef struct {U tymed;H handle;void*release;} Medium;
typedef struct {void*qi,*add,*release;HR(*get)(void*,Format*,Medium*);} DataVT;
_Static_assert(sizeof(Format)==32 && offsetof(Format,ptd)==8 && offsetof(Format,tymed)==24,"FORMATETC");
_Static_assert(sizeof(Medium)==24 && offsetof(Medium,handle)==8,"STGMEDIUM");
static HR init(Iface*self,const void*pidl,void*data,H key){(void)key;Obj*o=self->owner;o->count=o->available=0;o->path[0]=0;drop(o->data);o->data=0;o->exe=0;
 if(o->scope==0){if(!pidl||!SHGetPathFromIDListEx(pidl,o->path,MAXPATH,0))return FAIL;}
 else{if(!data)return FAIL;Format f={15,0,1,-1,1};Medium m={0,0,0};DataVT*vt=*(DataVT**)data;if(vt->get(data,&f,&m)<0)return FAIL;U got=0;if(m.tymed==1&&m.handle&&DragQueryFileW(m.handle,0xffffffff,0,0)==1){U need=DragQueryFileW(m.handle,0,0,0);if(need&&need<MAXPATH)got=DragQueryFileW(m.handle,0,o->path,MAXPATH);}ReleaseStgMedium(&m);if(!got)return FAIL;}
 if(!pathok(o->path))return FAIL;U n=len(o->path);while(n>3&&o->path[n-1]=='\\')o->path[--n]=0;
 U bytes=0;const W*keyname=WSTR("Software\\HoI4ModNavigator\\ShellV2");const W*val=WSTR("Menu");if(RegGetValueW(HKCU,keyname,val,2,0,0,&bytes)||bytes<4||bytes>MAXSNAP||(bytes&1))return OK;
 o->data=alloc((size_t)bytes+2);if(!o->data)return OOM;U read=bytes;if(RegGetValueW(HKCU,keyname,val,2,0,o->data,&read)||read>bytes||(read&1)||read<4)return OK;
 W*next=o->data,*end=o->data+read/2;W*ver=line(&next,end);W*exe=line(&next,end);if(!ver||!exe||!same(ver,WSTR("H4N2"))||!pathok(exe))return OK;o->exe=exe;
 int localBlocked=0;while(next<end&&*next){W*scope=line(&next,end),*root=line(&next,end),*label=line(&next,end),*action=line(&next,end),*id=line(&next,end);if(!scope||!root||!label||!action||!id){o->available=0;return OK;}if(scope[0]!=(W)('0'+o->scope)||scope[1]||!within(root,o->path)||!token(action)||!token(id))continue;
 if(same(action,WSTR("skip-local"))){localBlocked=1;continue;}
 if(same(id,WSTR("local"))&&(localBlocked||!localwithin(root,o->path,o->scope)))continue;
 if(!len(label))continue;if(o->available>=MAXCOMMANDS){o->available=0;return OK;}o->entries[o->available++]=(Entry){label,action,id};}
 return OK;
}
static HR query(Iface*self,H menu,U at,U first,U last,U flags){Obj*o=self->owner;o->count=0;if(flags&1||last<first)return OK;for(U i=0;i<o->available && i<=last-first;i++){Entry e=o->entries[i];if(!InsertMenuW(menu,at+o->count,0x400,first+o->count,e.label))break;o->shown[o->count++]=e;}return (HR)o->count;}
/* CMINVOKECOMMANDINFO / EX layouts */
typedef struct {U size,mask;H hwnd;const char*verb,*params,*dir;int32_t show;U hotkey;H icon;const char*title;const W*verbW,*paramsW,*dirW,*titleW;int32_t x,y;} Invoke;
_Static_assert(offsetof(Invoke,verb)==16 && offsetof(Invoke,title)==56 && offsetof(Invoke,verbW)==64 && sizeof(Invoke)==104+0,"CMINVOKECOMMANDINFOEX");
static W* append(W*p,const W*s){while(*s)*p++=*s++;*p=0;return p;}
static U canonical(Entry e,W*out){W*p=append(out,WSTR("hoi4nav."));p=append(p,e.action);p=append(p,WSTR("."));p=append(p,e.id);return (U)(p-out);}
static HR getstring(Iface*self,uintptr_t id,U flags,U*reserved,char*out,U cap){(void)reserved;Obj*o=self->owner;if(id>=o->count)return INVALID;if(flags==2||flags==6)return OK;if(flags!=0&&flags!=4)return NOTIMPL;if(!out||!cap)return INVALID;W b[96];U n=canonical(o->shown[id],b);if(n>=cap)return INVALID;if(flags==4){W*w=(W*)out;for(U i=0;i<=n;i++)w[i]=b[i];}else for(U i=0;i<=n;i++)out[i]=(char)b[i];return OK;}
static HR invoke(Iface*self,const Invoke*v){Obj*o=self->owner;if(!v||v->size<56||!o->exe)return INVALID;uintptr_t raw=(uintptr_t)v->verb;U id=MAXCOMMANDS;
 if(raw<=0xffff)id=(U)raw;else{for(U i=0;i<o->count;i++){W b[96];U n=canonical(o->shown[i],b);int match=1;if(v->size>=sizeof(Invoke)&&(v->mask&0x4000)&&v->verbW&&(uintptr_t)v->verbW>0xffff)match=same(b,v->verbW);else{for(U j=0;j<=n;j++)if((unsigned char)v->verb[j]!=b[j]){match=0;break;}}if(match){id=i;break;}}}
 if(id>=o->count)return INVALID;Entry e=o->shown[id];W*args=alloc((len(o->path)+len(e.action)+len(e.id)+20)*sizeof(W));if(!args)return OOM;W*p=append(args,WSTR("action "));p=append(p,e.action);p=append(p,WSTR(" "));p=append(p,e.id);p=append(p,WSTR(" \""));p=append(p,o->path);append(p,WSTR("\""));H r=ShellExecuteW(v->hwnd,WSTR("open"),o->exe,args,0,1);drop(args);return (uintptr_t)r>32?OK:FAIL;
}
static const struct {HR(*qi)(Iface*,const GUID*,void**);U(*add)(Iface*),(*release)(Iface*);HR(*init)(Iface*,const void*,void*,H);} initVT={qi,add,release,init};
static const struct {HR(*qi)(Iface*,const GUID*,void**);U(*add)(Iface*),(*release)(Iface*);HR(*query)(Iface*,H,U,U,U,U);HR(*invoke)(Iface*,const Invoke*);HR(*string)(Iface*,uintptr_t,U,U*,char*,U);} menuVT={qi,add,release,query,invoke,getstring};
typedef struct {const void*vt;long ref;U scope;} Factory;
static HR fqi(Factory*f,const GUID*i,void**out){if(!out)return INVALID;*out=0;if(!eqg(i,&iidUnknown)&&!eqg(i,&iidFactory))return NOIF;*out=f;inc(&f->ref);return OK;}
static U fadd(Factory*f){return inc(&f->ref);}
static U frelease(Factory*f){U n=dec(&f->ref);if(!n){drop(f);dec(&live);}return n;}
static HR create(Factory*f,void*outer,const GUID*i,void**out){if(!out)return INVALID;*out=0;if(outer)return NOAGG;Obj*o=alloc(sizeof(Obj));if(!o)return OOM;inc(&live);o->ref=1;o->scope=f->scope;o->init=(Iface){&initVT,o};o->menu=(Iface){&menuVT,o};HR hr=qi(&o->init,i,out);release(&o->init);return hr;}
static HR lock(Factory*f,int on){(void)f;if(on)inc(&locks);else dec(&locks);return OK;}
static const struct {HR(*qi)(Factory*,const GUID*,void**);U(*add)(Factory*),(*release)(Factory*);HR(*create)(Factory*,void*,const GUID*,void**);HR(*lock)(Factory*,int);} factoryVT={fqi,fadd,frelease,create,lock};
__declspec(dllexport) HR DllGetClassObject(const GUID*c,const GUID*i,void**out){if(!out)return INVALID;*out=0;U scope=0;while(scope<3&&!eqg(c,&classes[scope]))scope++;if(scope==3)return CLASSNA;Factory*f=alloc(sizeof(Factory));if(!f)return OOM;inc(&live);f->vt=&factoryVT;f->ref=1;f->scope=scope;HR hr=fqi(f,i,out);frelease(f);return hr;}
__declspec(dllexport) HR DllCanUnloadNow(void){return __atomic_load_n(&live,__ATOMIC_SEQ_CST)||__atomic_load_n(&locks,__ATOMIC_SEQ_CST)?NO:OK;}
