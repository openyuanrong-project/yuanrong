/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.function;

import org.yuanrong.api.InvokeArg;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.function.common.Util;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.InvokeOptions;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.services.model.CallRequest;
import org.yuanrong.services.runtime.Context;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.Locale;

/**
 * The type Function.
 *
 * @since 2024/8/15
 */
public class Function {
    private static Logger LOG = LoggerFactory.getLogger(Function.class);

    private final Context context;

    private String functionNameWithVersion;

    private String functionId;

    private CreateOptions createOptions = new CreateOptions();

    private String instanceId;

    /**
     * construct method
     *
     * @param context context
     */
    public Function(Context context) {
        this(context, null);
    }

    /**
     * construct method
     *
     * @param functionNameWithVersion functionNameWithVersion, eg: javafunc:latest
     */
    public Function(String functionNameWithVersion) {
        this(null, functionNameWithVersion);
    }

    /**
     * construct method
     *
     * @param context context
     * @param functionNameWithVersion functionNameWithVersion, eg: javafunc:latest
     */
    public Function(Context context, String functionNameWithVersion) {
        this.context = context;
        this.functionNameWithVersion = functionNameWithVersion;
    }

    /**
     * options
     *
     * @param opts the CreateOptions value
     * @return Function
     */
    public Function options(CreateOptions opts) {
        Util.checkDynamicResource(opts.getCpu(), opts.getMemory());
        this.createOptions = opts;
        return this;
    }

    /**
     * SDK invoke api
     *
     * @param payload passed argument
     * @return ObjectRef
     */
    public <T> ObjectRef<T> invoke(String payload) {
        LOG.info("SDK invoke api beginning");
        String functionService = org.yuanrong.services.runtime.utils.Util.getServiceNameFromEnv(context);
        String tenantId = org.yuanrong.services.runtime.utils.Util.getTenantIdFromEnv(context);
        String[] nameAndVersion = Util.checkFuncName(this.functionNameWithVersion);
        this.functionId = String.format(Locale.ROOT, "%s/0@%s@%s/%s", tenantId, functionService, nameAndVersion[0],
            nameAndVersion[1]);

        FunctionMeta funcMeta = FunctionMeta.newBuilder()
            .setApiType(ApiType.Faas)
            .setFunctionID(this.functionId)
            .setLanguage(LanguageType.Java)
            .build();
        Util.checkPayload(payload);
        CallRequest arg1 = new CallRequest.Builder().build();
        CallRequest arg2 = new CallRequest.Builder().withBody(payload).build();
        List<InvokeArg> argList;
            argList = Utils.packFaasInvokeArgs(arg1, arg2);
        Pair<ErrorInfo, String> res;
        try {
            InvokeOptions invokeOptions = createOptions.convertInvokeOptions();
            if (context != null) {
                invokeOptions.setTraceId(context.getTraceID());
            }
            res = LibRuntime.InvokeInstance(funcMeta, "", argList, invokeOptions);
        } catch (LibRuntimeException e) {
            throw new InvokeException(e.getErrorCode().getValue(), e.getMessage());
        }
        Util.checkErrorAndThrow(res.getFirst(), "faas invoke");
        return new ObjectRef<>(res.getSecond());
    }

    /**
     * SDK terminate api
     *
     * @return ObjectRef
     * @throws InvokeException when internal error
     */
    public ObjectRef terminate() {
        LOG.debug("SDK terminate api beginning");
        return null;
    }

    /**
     * SDK saveState api
     *
     * @throws InvokeException when internal error
     */
    public void saveState() {
        LOG.debug("saveState api beginning");
    }

    /**
     * get context
     *
     * @return context
     */
    public Context getContext() {
        return this.context;
    }

    /**
     * SDK getInstanceID api
     *
     * @return InstanceID
     */
    public String getInstanceID() {
        return this.instanceId;
    }
}
