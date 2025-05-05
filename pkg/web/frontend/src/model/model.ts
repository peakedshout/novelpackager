export class SearchResult {
    name: string = "";
    id: string = "";
    author: string = "";
    cover: string = "";
    description: string = "";
    ahref: string = "";
    metas: string[] = [];
}

export class BookInfo {
    name: string = "";
    id: string = "";
    author: string = "";
    cover: string = "";
    description: string = "";
    metas: string[] = [];
    volumes: VolumeInfo[] = [];
}

export class VolumeInfo {
    name: string = "";
    id: string = "";
    cover: string = "";
    description: string = "";
    chapters: ChapterInfo[] = [];
}

export class ChapterInfo {
    name: string = "";
}
