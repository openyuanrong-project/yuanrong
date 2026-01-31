import { DownloadLogPath } from '@/api/api';
import { WarningNotify } from '@/components/warning-notify';

export function DownloadLogHandler(filename: string) {
    const controller = new AbortController();
    const { signal } = controller;
    fetch(DownloadLogPath + filename, { signal }).then((res) => {
        if (res.ok) {
            controller.abort();
            window.location.href = DownloadLogPath + filename;
        } else {
            res.json().then((data)=>{
                throw data?.message;
            }).catch((err: any)=>{
                WarningNotify('DownloadLogByFilenameAPI', err);
            });
        }
    })
}