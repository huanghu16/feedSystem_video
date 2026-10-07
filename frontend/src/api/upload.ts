import { postJson, postFormLong } from './client'

const CHUNK_SIZE = 5 * 1024 * 1024 // 5MB，需与后端 maxChunkSize 一致
const MAX_RETRY = 3

export interface UploadProgress {
    uploadedBytes: number
    totalBytes: number
    percent: number
}

interface InitResponse {
    upload_id: string
    uploaded_chunks: number[] | null
    instant: boolean
    play_url?: string
    cover_url?: string
}

interface MergeResponse {
    play_url: string
    cover_url: string
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

/**
 * 计算文件 sha256，作为 upload_id
 * 注意：crypto.subtle 只在安全上下文可用（HTTPS 或 localhost）
 */
async function calcFileHash(file: File): Promise<string> {
    if (!window.crypto?.subtle) {
        throw new Error('当前环境不支持文件校验，请使用 HTTPS 或 localhost 访问')
    }
    const buf = await file.arrayBuffer()
    const digest = await crypto.subtle.digest('SHA-256', buf)
    return Array.from(new Uint8Array(digest))
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('')
}

/** 第 i 片实际字节数（最后一片通常不满） */
function chunkSizeAt(file: File, index: number): number {
    return Math.min(CHUNK_SIZE, file.size - index * CHUNK_SIZE)
}

async function requestInit(file: File, hash: string, totalChunks: number): Promise<InitResponse> {
    return postJson<InitResponse>('/video/upload/init', {
        file_hash: hash,
        file_name: file.name,
        file_size: file.size,
        chunk_size: CHUNK_SIZE,
        total_chunks: totalChunks,
    })
}

async function uploadOneChunk(hash: string, index: number, blob: Blob): Promise<void> {
    const fd = new FormData()
    fd.append('upload_id', hash)
    fd.append('chunk_index', String(index))
    fd.append('file', blob)
    await postFormLong('/video/upload/chunk', fd)
}

/** 单分片上传，失败按 500ms/1s/2s 退避重试 */
async function uploadChunkWithRetry(hash: string, index: number, blob: Blob): Promise<void> {
    let lastErr: any
    for (let attempt = 1; attempt <= MAX_RETRY; attempt++) {
        try {
            await uploadOneChunk(hash, index, blob)
            return
        } catch (err: any) {
            lastErr = err
            // 4xx 是客户端问题（参数非法等），重试没意义
            if (err?.status >= 400 && err?.status < 500) throw err
            if (attempt < MAX_RETRY) {
                console.warn(`分片 ${index} 第 ${attempt} 次失败，准备重试`, err)
                await sleep(500 * 2 ** (attempt - 1))
            }
        }
    }
    throw lastErr
}

/**
 * 分片上传视频（带断点续传 + 秒传）
 * @param file 视频文件
 * @param onProgress 进度回调
 */
export async function uploadVideoChunked(
    file: File,
    onProgress?: (p: UploadProgress) => void
): Promise<MergeResponse> {
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE)
    const hash = await calcFileHash(file)

    // 1. 初始化：拿 upload_id 和已上传分片列表
    const init = await requestInit(file, hash, totalChunks)

    // 命中秒传，文件已在服务端
    if (init.instant) {
        onProgress?.({ uploadedBytes: file.size, totalBytes: file.size, percent: 100 })
        return { play_url: init.play_url || '', cover_url: init.cover_url || '' }
    }

    const uploaded = new Set(init.uploaded_chunks || [])
    let uploadedBytes = 0
    uploaded.forEach((i) => {
        uploadedBytes += chunkSizeAt(file, i)
    })

    const report = () => {
        onProgress?.({
            uploadedBytes,
            totalBytes: file.size,
            percent: Math.min(100, Math.round((uploadedBytes / file.size) * 100)),
        })
    }
    report()

    // 2. 逐片上传，已存在的直接跳过 —— 这就是断点续传
    for (let i = 0; i < totalChunks; i++) {
        if (uploaded.has(i)) continue

        const start = i * CHUNK_SIZE
        const blob = file.slice(start, Math.min(start + CHUNK_SIZE, file.size))

        await uploadChunkWithRetry(hash, i, blob)

        uploadedBytes += blob.size
        report()
    }

    // 3. 合并 + 服务端校验 + 生成封面
    return postJson<MergeResponse>('/video/upload/merge', {
        upload_id: hash,
        file_name: file.name,
        total_chunks: totalChunks,
    })
}