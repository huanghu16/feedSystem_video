import { postJson } from './client'

/**
 * 关注用户请求参数
 * @param vloggerId 要关注的用户id
 * @returns 关注结果
 */
export interface FollowRequest {
    vlogger_id: number
}

/**
 * 取消关注请求参数
 * @param vloggerId 要取消关注的用户id
 * @returns 取消关注结果
 */
export interface UnfollowRequest {
    vlogger_id: number
}

/**
 * 检查是否已关注
 * @param vloggerId 用户id
 * @returns 是否已关注
 */
export async function isFollowing(vloggerId: number): Promise<{ is_following: boolean }> {
    return await postJson('/social/isFollowing', { vlogger_id: vloggerId })
}

/**
 * 关注用户
 * @param vloggerId 要关注的用户id
 * @returns 关注结果
 */
export async function follow(vloggerId: number): Promise<{ message: string }> {
    return await postJson('/social/follow', { vlogger_id: vloggerId })
}

/**
 * 取消关注
 * @param vloggerId 要取消关注的用户id
 * @returns 取消关注结果
 */
export async function unfollow(vloggerId: number): Promise<{ message: string }> {
    return await postJson('/social/unfollow', { vlogger_id: vloggerId })
}
