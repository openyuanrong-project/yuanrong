/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Float2 } from '../utils/handleNum';

const percentageBarStyle = {
    background: "#ffffff",
    border: "1px solid #e6f4ff",
    minHeight: "14px",
    lineHeight: "14px",
    position: "relative",
    boxSizing: "border-box",
    borderRadius: "2px",
}

const progressBarStyle = (percentage: number) => ({
    background: "#e6f4ff",
    position: "absolute",
    left: 0,
    minHeight: "14px",
    transition: "0.5s width",
    boxSizing: "border-box",
    width: `${Math.min(Math.max(0, percentage), 100)}%`,
})

const textStyle = {
    fontSize: 14,
    position: "relative",
    width: "100%",
    textAlign: "center",
    whiteSpace: "nowrap",
}

export const ProgressBar = (h:any, used: number, total: number) => {
    const percent = total ? Float2(used / total * 100) : 0;
    return h('div', {style: percentageBarStyle}, [
        h('div', {style: progressBarStyle(percent)}),
        h('div', {style: textStyle}, used + '/' + total + '(' + percent + '%)'),
    ]);
}

export const SimpleProgressBar = (h:any, percent: number) => {
    percent = Float2(percent);
    return h('div', {style: percentageBarStyle}, [
        h('div', {style: progressBarStyle(percent)}),
        h('div', {style: textStyle}, percent + '%'),
    ]);
}

export const MultiProgressBar = (h:any, data: object) => {
    let children = [];
    Object.keys(data).forEach(key => {
        const percent = data[key].capacity ? Float2(data[key].used / data[key].capacity * 100) : 0;
        const child = h('div', [
            h('span', ["[" + key + "]:"]),
            h('div',
                {style: {...percentageBarStyle, marginLeft: '4px', marginBottom: '4px',
                        display: 'inline-block', width: '80%',}},
                [
                    h('div', {style: progressBarStyle(percent)}),
                    h('div', {style: textStyle}, percent + '%'),
                ]
            ),
        ]);
        children.push(child);
    })
    return h('div', children);
}