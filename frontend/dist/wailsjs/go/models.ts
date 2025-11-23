export namespace main {
	
	export class EmojiData {
	    emoji: string;
	    key: string;
	
	    static createFrom(source: any = {}) {
	        return new EmojiData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.emoji = source["emoji"];
	        this.key = source["key"];
	    }
	}

}

