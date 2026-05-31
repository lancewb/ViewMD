export namespace models {
	
	export class WorkspaceState {
	    projectPath: string;
	    openFiles: string[];
	    activeFile: string;
	    viewMode: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.openFiles = source["openFiles"];
	        this.activeFile = source["activeFile"];
	        this.viewMode = source["viewMode"];
	    }
	}
	export class ProjectSummary {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class AppState {
	    lastProject?: ProjectSummary;
	    workspace: WorkspaceState;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastProject = this.convertValues(source["lastProject"], ProjectSummary);
	        this.workspace = this.convertValues(source["workspace"], WorkspaceState);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TreeNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    isMarkdown: boolean;
	    size: number;
	    modifiedTime: string;
	    children?: TreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.isMarkdown = source["isMarkdown"];
	        this.size = source["size"];
	        this.modifiedTime = source["modifiedTime"];
	        this.children = this.convertValues(source["children"], TreeNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileDocument {
	    name: string;
	    path: string;
	    absolutePath: string;
	    content: string;
	    size: number;
	    modifiedTime: string;
	
	    static createFrom(source: any = {}) {
	        return new FileDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.absolutePath = source["absolutePath"];
	        this.content = source["content"];
	        this.size = source["size"];
	        this.modifiedTime = source["modifiedTime"];
	    }
	}
	export class CreatedFileResult {
	    document: FileDocument;
	    tree: TreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new CreatedFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.document = this.convertValues(source["document"], FileDocument);
	        this.tree = this.convertValues(source["tree"], TreeNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ProjectSnapshot {
	    project: ProjectSummary;
	    tree: TreeNode[];
	    openDocuments: FileDocument[];
	    activeFile: string;
	    viewMode: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], ProjectSummary);
	        this.tree = this.convertValues(source["tree"], TreeNode);
	        this.openDocuments = this.convertValues(source["openDocuments"], FileDocument);
	        this.activeFile = source["activeFile"];
	        this.viewMode = source["viewMode"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

