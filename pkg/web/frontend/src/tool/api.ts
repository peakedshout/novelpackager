import {type MsgContainer} from "./tool1.ts";
import type {BookInfo, SearchResult} from '../model/model.ts'
import type {Error} from "./err.ts";

export class Api {
    async failedFunc(res: Response) {
        const data = await res.text()
        if (data.length > 0) {
            return Promise.reject(data)
        } else {
            return Promise.reject(res.statusText)
        }
    }

    async GetSourceList(): Promise<string[]> {
        const res = await fetch('/api/source_list')
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async Search(source: string, name: string): Promise<MsgContainer<SearchResult[]>> {
        const url = new URL('/api/search', window.location.origin);
        url.searchParams.append('source', source);
        url.searchParams.append('name', name);
        url.searchParams.append('full', false.toString());
        const res = await fetch(url.toString())
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async GetBookInfo(source: string, id: string): Promise<MsgContainer<BookInfo>> {
        const url = new URL('/api/get_info', window.location.origin);
        url.searchParams.append('source', source);
        url.searchParams.append('id', id);
        url.searchParams.append('full', true.toString());
        const res = await fetch(url.toString())
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async GetProgress(source: string): Promise<MsgContainer<Map<string, string>>> {
        const url = new URL('/api/progress', window.location.origin);
        url.searchParams.append('source', source);
        const res = await fetch(url.toString())
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async GetEnableDownload(source: string, id: string): Promise<MsgContainer<string[]>> {
        const url = new URL('/api/enable_download', window.location.origin);
        url.searchParams.append('source', source);
        url.searchParams.append('id', id);
        const res = await fetch(url.toString())
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async Caching(source: string, id: string): Promise<Error> {
        const url = new URL('/api/caching', window.location.origin);
        url.searchParams.append('source', source);
        url.searchParams.append('id', id);
        const res = await fetch(url.toString())
        if (!res.ok) {
            await this.failedFunc(res)
        }
        return await res.json()
    }

    async Download(source: string, id: string, vols: number[]) {
        const url = new URL('/api/download', window.location.origin);
        url.searchParams.append('source', source);
        url.searchParams.append('id', id);
        url.searchParams.append('vols', vols.join(','))
        const res = await fetch(url);
        if (!res.ok) {
            await this.failedFunc(res)
        }
        const blob = await res.blob();
        const disposition = res.headers.get("Content-Disposition");
        const filename = getFileNameFromDisposition(disposition);
        const downloadUrl = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = downloadUrl
        a.download = filename
        document.body.appendChild(a)
        a.click()
        a.remove()
        URL.revokeObjectURL(downloadUrl)
    }
}

export const api = new Api()

function getFileNameFromDisposition(disposition: string | null): string {
    if (!disposition) return 'download.epub';

    let filename = 'download.epub';

    // Try to match RFC 5987 format (filename*=)
    const filenameStarMatch = disposition.match(/filename\*\s*=\s*([^']*)''([^;]+)/i);
    if (filenameStarMatch) {
        try {
            filename = decodeURIComponent(filenameStarMatch[2]);
            return filename;
        } catch (e) {
            console.warn('Failed to decode filename*:', e);
        }
    }

    // Fallback to basic filename=
    const filenameMatch = disposition.match(/filename\s*=\s*"?([^"]+)"?/i);
    if (filenameMatch) {
        filename = filenameMatch[1];
    }

    return filename;
}
