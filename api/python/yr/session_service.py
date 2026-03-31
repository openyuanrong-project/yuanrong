#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Session Service Module

Provides capabilities for loading and modifying Session.
Session creation, saving, and lock management are all automatically handled by runtime.
SDK is only responsible for reading and modifying.
"""

import json
from typing import List, Optional

from yr import log


class ManagedSessionObj:
    """
    Managed Session Object.

    Session Data Structure:
        {
            "sessionID": "s-123",
            "histories": ["user: hello", "assistant: hi"]
        }

    Semantic Constraints:
    - Internally holds sessionId and the latest sessionData JSON string
    - get_histories() returns a copy to prevent users from directly modifying internal list and bypassing sync
    - set_histories() immediately calls bridge to sync to libruntime, runtime auto-persists at request end
    - Multiple load_session() calls with same sessionId return the same managed object
    """

    def __init__(self, session_id: str, session_json: str):
        """
        Construct managed Session Object.

        :param session_id: Session ID
        :param session_json: JSON string of Session data
        """
        self._session_id = session_id
        self._session_json = session_json
        self._data = json.loads(session_json) if session_json else {}
        self._histories = self._data.get("histories", [])

    def get_id(self) -> str:
        """
        Get Session ID.

        Returns:
            Session ID string
        """
        return self._session_id

    def get_histories(self) -> List[str]:
        """
        Get history list (copy).

        Returns a snapshot copy, should not expose as writable internal reference.
        After user modification, must call set_histories() to write back,
        so SDK can sync changes to libruntime.

        Returns:
            Shallow copy of history list
        """
        return list(self._histories)

    def set_histories(self, histories: List[str]) -> None:
        """
        Set history list and immediately sync to libruntime.

        1. Update local in-memory histories
        2. Construct the latest session JSON
        3. Call bridge UpdateCurrentSession to sync to libruntime
        4. At request end, runtime will auto-persist session to DataSystem

        :param histories: New history list
        :raises RuntimeError: raised when bridge call fails
        """
        self._histories = histories
        self._data["histories"] = histories
        self._session_json = json.dumps(self._data)

        try:
            from yr.fnruntime import update_current_session
            update_current_session(self._session_id, self._session_json)
        except Exception as e:
            log.get_logger().error(
                f"Failed to update current session, sessionId={self._session_id}, err={e}")
            raise RuntimeError(
                f"Failed to update session, err: {e}"
            ) from e


class SessionService:
    """
    Session Service Interface.

    Provides load_session() method to load Session object for current invocation from libruntime memory.
    Does not directly query DataSystem, but calls libruntime's LoadCurrentSession via bridge,
    libruntime reads the currently active AgentSessionContext from activeSessionMap.
    """

    def __init__(self, session_id: str):
        """
        Construct SessionService.

        :param session_id: Session ID of current invocation (passed from Context)
        """
        self._session_id = session_id

    def load_session(self) -> Optional["ManagedSessionObj"]:
        """
        Load Session object for current invocation.

        - Uses Context.sessionId to call native LoadCurrentSession(sessionId)
        - libruntime reads currently active AgentSessionContext from activeSessionMap[sessionId]
        - Returns language wrapper of in-memory Session object

        Semantics:
        - Returns None when current request does not carry sessionId
        - Returns None when use_agent_session=false
        - When use_agent_session=true and session is not created, runtime automatically creates
          empty SessionObj and writes to DataSystem, load_session() still returns object (with empty content)

        Returns:
            ManagedSessionObj, or None when sessionId is empty

        Raises:
            RuntimeError: raised when bridge call fails
        """
        if not self._session_id:
            log.get_logger().debug("Session ID is empty, skip load_session.")
            return None

        try:
            from yr.fnruntime import load_current_session
            session_json = load_current_session(self._session_id)
            if session_json is None or session_json == "":
                log.get_logger().warning(
                    f"load_session returned empty, sessionId={self._session_id}")
                return None
            return ManagedSessionObj(self._session_id, session_json)
        except Exception as e:
            log.get_logger().error(
                f"Failed to load session, sessionId={self._session_id}, err={e}")
            raise RuntimeError(
                f"Failed to load session, sessionId={self._session_id}, err: {e}"
            ) from e
