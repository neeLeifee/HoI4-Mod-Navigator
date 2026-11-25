/* Host-side mock tests: COM dispatch + menu parsing, not Windows integration. */
#define __declspec(x)
#include "shell.c"
#include <assert.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
static const W* snap;
static U selectedCount=1;
static const W* selectedPath;
static U menuCount;
static const W* labels[80];
static W launched[1024];
static U launches;
H GetProcessHeap(void){return (H)1;}
H HeapAlloc(H h,U f,size_t n){(void)h;(void)f;return calloc(1,n);}
int HeapFree(H h,U f,H p){(void)h;(void)f;free(p);return 1;}
static W fold(W w){return w>='A'&&w<='Z'?w+32:w;}
int CompareStringOrdinal(const W*a,int an,const W*b,int bn,int ignore){if(an<0)an=len(a);if(bn<0)bn=len(b);if(an!=bn)return 1;for(int i=0;i<an;i++)if((ignore?fold(a[i]):a[i])!=(ignore?fold(b[i]):b[i]))return 1;return 2;}
int32_t RegGetValueW(H h,const W*k,const W*v,U flags,U*t,void*out,U*n){(void)h;(void)k;(void)v;(void)flags;(void)t;U need=(len(snap)+1)*2;if(!out){*n=need;return 0;}if(*n<need)return 234;memcpy(out,snap,need);*n=need;return 0;}
int SHGetPathFromIDListEx(const void*p,W*out,U n,U flags){(void)flags;if(!p||len(p)>=n)return 0;append(out,p);return 1;}
U DragQueryFileW(H h,U i,W*out,U n){(void)h;if(i==0xffffffff)return selectedCount;if(!out)return len(selectedPath);if(len(selectedPath)>=n)return 0;append(out,selectedPath);return len(selectedPath);}
void ReleaseStgMedium(void*p){(void)p;}
int InsertMenuW(H h,U at,U flags,uintptr_t id,const W*label){(void)h;(void)at;(void)flags;(void)id;assert(menuCount<80);labels[menuCount++]=label;return 1;}
H ShellExecuteW(H h,const W*verb,const W*exe,const W*args,const W*dir,int show){(void)h;(void)verb;(void)dir;(void)show;assert(same(exe,WSTR("C:\\Nav.exe")));assert(len(args)<1024);append(launched,args);launches++;return (H)33;}
static HR mockGet(void*o,Format*f,Medium*m){(void)o;assert(f->format==15&&f->aspect==1&&f->tymed==1&&f->index==-1);m->tymed=1;m->handle=(H)1;return OK;}
static DataVT dataVT={0,0,0,mockGet};static DataVT* data=&dataVT;
static void scenario(U scope,const W*path,U count,const W*label,const W*args){
 Factory*f=0;Iface*i=0,*m=0;menuCount=0;launches=0;
 assert(DllGetClassObject(&classes[scope],&iidFactory,(void**)&f)==OK);assert(DllCanUnloadNow()==NO);
 assert(create(f,0,&iidInit,(void**)&i)==OK);assert(qi(i,&iidMenu,(void**)&m)==OK);assert(frelease(f)==0);
 selectedPath=path;HR hr=init(i,scope?0:path,scope?&data:0,0);
 if(scope&&selectedCount!=1)assert(hr<0);else assert(hr==OK);
 assert(query(m,0,0,1,100,1)==0);assert(menuCount==0);assert(query(m,0,0,1,100,0)==(HR)count);assert(menuCount==count);
 if(count){assert(same(labels[0],label));Invoke v={0};v.size=56;v.verb=0;assert(invoke(m,&v)==OK);assert(launches==1&&same(launched,args));v.verb=(void*)(uintptr_t)count;assert(invoke(m,&v)==INVALID);char name[96];assert(getstring(m,0,0,0,name,96)==OK);v.verb=name;assert(invoke(m,&v)==OK);assert(launches==2);}
 assert(release(i)==1);assert(release(m)==0);assert(DllCanUnloadNow()==OK);
}
int main(void){
 snap=WSTR("H4N2\nC:\\Nav.exe\n0\nC:\\Game\nOpen FFU2\nopen\n123\n0\nD:\\Mods\\FFU2\nOpen vanilla\nreverse\n123\n1\nC:\\Game\nOpen folder FFU2\nopen\n123\n2\nD:\\Mods\\FFU2\nSelect original\nreverse-file\n123\n2\nC:\\Game\nCopy FFU2\ncopy\n123\n");
 scenario(0,WSTR("c:\\game\\gfx"),1,WSTR("Open FFU2"),WSTR("action open 123 \"c:\\game\\gfx\""));
 scenario(0,WSTR("D:\\Mods\\FFU2"),1,WSTR("Open vanilla"),WSTR("action reverse 123 \"D:\\Mods\\FFU2\""));
 scenario(0,WSTR("D:\\Mods\\FFU2-backup"),0,0,0);
 scenario(0,WSTR("C:\\Other"),0,0,0);
 scenario(1,WSTR("C:\\Game\\gfx"),1,WSTR("Open folder FFU2"),WSTR("action open 123 \"C:\\Game\\gfx\""));
 scenario(2,WSTR("D:\\Mods\\FFU2\\a.txt"),1,WSTR("Select original"),WSTR("action reverse-file 123 \"D:\\Mods\\FFU2\\a.txt\""));
 scenario(2,WSTR("C:\\Game\\a.txt"),1,WSTR("Copy FFU2"),WSTR("action copy 123 \"C:\\Game\\a.txt\""));
 selectedCount=2;scenario(2,WSTR("C:\\Game\\a.txt"),0,0,0);
 selectedCount=1;snap=WSTR("H4N2\nC:\\Nav.exe\n0\nC:\\Game\nBroken\n");scenario(0,WSTR("C:\\Game"),0,0,0);
 puts("9 mock scenarios passed");return 0;
}
