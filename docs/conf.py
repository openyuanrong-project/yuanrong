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

import sys
import os
import logging
import datetime
from pathlib import Path

sys.path.append(os.path.abspath('.'))
from icons import ICONS
from blog_data import BLOG_DATA

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)

if os.getenv("BUILD_WITH_PACKAGE", "").lower() == "false":
    sys.path.insert(0, str(Path("..", "api", "python").resolve()))
else:
    logging.info("The installed openYuanrong package will be used to generate the Python API doc")


ENV_YR_GIT_COMMIT_ID = os.environ.get("YR_DOC_GIT_COMMIT_ID", "")
ENV_BUILD_VERSION = os.environ.get("BUILD_VERSION", "")

build_time = datetime.datetime.now(tz=datetime.timezone.utc) + datetime.timedelta(
    hours=8
)
current_time_str = build_time.strftime("%Y-%m-%d %H:%M:%S")

project = "openYuanrong"
copyright = f"{build_time.year}, openEuler openYuanrong"
author = "openYuanrong with CC BY 4.0 LICENSE"

logging.info(
    f"""Doc build configs:
ENV_YR_GIT_COMMIT_ID: {ENV_YR_GIT_COMMIT_ID}
ENV_BUILD_VERSION: {ENV_BUILD_VERSION}
current_date: {current_time_str}
project: {project}
copyright: {copyright}
author: {author}
"""
)

templates_path = ["_templates"]
exclude_patterns = [
    "_build",
    "README.md",
    "sample_code",
    "multi_language_function_programming_interface/api/distributed_programming/C++",
    "multi_language_function_programming_interface/api/distributed_programming/Java",
    "multi_language_function_programming_interface/api/distributed_programming/Python",
    "multi_language_function_programming_interface/api/function_service/Java",
    "multi_language_function_programming_interface/api/function_service/Cpp",
    "multi_language_function_programming_interface/development_guide/function_service/response-stream.md",

    ""
]

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = [
    "sphinx.ext.autodoc",
    "sphinx.ext.autosummary",
    "sphinx.ext.napoleon",
    "sphinx.ext.viewcode",
    "sphinx_design",
    "sphinx_copybutton",
    "sphinx_togglebutton",
    "myst_parser",
    "breathe",
    "sphinxcontrib.openapi",  # 添加 sphinxcontrib-openapi 扩展
]

autoclass_content = "both"
copybutton_exclude = ".linenos, .gp, .go"


# -----------------------------------------------------------------------------
#   FOR PYTHON API GENEREATE
# -----------------------------------------------------------------------------
autodoc_mock_imports = ["acl", "requests", "fastapi", "numpy"]

autosummary_generate = True
autosummary_generate_overwrite = True  # 覆盖已生成的文件
autosummary_ignore_module_all = False  # 不忽略 __all__ 的限制
autosummary_imported_members = True


# -----------------------------------------------------------------------------
#   HTML templates
# -----------------------------------------------------------------------------
html_logo = "./images/logo-small.png"
html_theme = "sphinx_book_theme"

html_static_path = ["_static"]
html_favicon = "_static/favicon.ico"
html_css_files = [
    "custom.css",
]

html_theme_options = {
    "show_navbar_depth": 1,
    "max_navbar_depth": 7,
    "collapse_navigation": True,
    "home_page_in_toc": True,
    "check_switcher": False,
    "extra_footer": """
        Built with
        <a href="https://www.sphinx-doc.org/en/master/">Sphinx</a>
        using a
        <a href="https://github.com/executablebooks/sphinx-book-theme">theme</a>
        provided by
        <a href="https://github.com/executablebooks">Executable Books Project</a>.
    """,
    "navbar_start": [
        "navbar-logo",
        "navbar-nav",
         ],
    "navbar_end": [
        "version-switcher",
        ],
    "switcher": {
        "json_url": "https://docs.openyuanrong.org/versions.json",
        "version_match": os.getenv("BUILD_VERSION", "latest"),
    },
    "logo": {
        "image_light": "_static/image-light.png",
        "image_dark": "_static/image-dark.png"
    },
    "secondary_sidebar_items": {
        "**": ["page-toc"],
        "index": []
    },
}

html_sidebars = {
    "**": ["search-button-field.html", "sbt-sidebar-nav.html"]
}

# 使用自定义首页模板
html_additional_pages = {
    'index': 'custom-index.html'
}

# -----------------------------------------------------------------------------
#   Myst extensions
# -----------------------------------------------------------------------------
myst_enable_extensions = [
    "dollarmath",
    "amsmath",
    "deflist",
    "fieldlist",
    "html_admonition",
    "html_image",
    "colon_fence",
    "smartquotes",
    "replacements",
    "strikethrough",
    "substitution",
    "tasklist",
    "attrs_inline",
    "attrs_block",
]

# -----------------------------------------------------------------------------
#   breathe config
# -----------------------------------------------------------------------------
breathe_projects = {"openYuanrong": "./../.doxygendocs/xml"}
breathe_default_project = "openYuanrong"


html_context = {
    **ICONS,
    "blog_data": BLOG_DATA,
    "metatags": """
            <meta name="author" content="openYuanrong Team">
            <meta name="keywords" content="openYuanrong, 分布式计算引擎, AI推理, Serverless">
            <meta name="description" content="openYuanrong 是业界领先的 Serverless 分布式计算引擎，致力于以一套统一 Serverless 架构支持 AI、大数据、微服务等各类分布式应用。">
        """
}
