export namespace config {
	
	export class Config {
	    theme: string;
	    sshKeyPaths: string[];
	    ignoreDirs: string[];
	    lastOpenedPath: string;
	    fontSize: number;
	    sidebarWidth: number;
	    recentPaths: string[];
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.sshKeyPaths = source["sshKeyPaths"];
	        this.ignoreDirs = source["ignoreDirs"];
	        this.lastOpenedPath = source["lastOpenedPath"];
	        this.fontSize = source["fontSize"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.recentPaths = source["recentPaths"];
	    }
	}

}

export namespace types {
	
	export class FileNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    children?: FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], FileNode);
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

