認可サーバー実装 (Go)
=====================

概要
----

[OAuth 2.0][RFC6749] と [OpenID Connect][OIDC] をサポートする認可サーバーの Go
による実装です。

この実装は Gin API と authlete-go-gin ライブラリを用いて書かれています。
[Gin][Gin] は Go で書かれた Web フレームワークの一つです。 [authlete-go-gin][AuthleteGoGin]
は、認可サーバーとリソースサーバーを実装するためのユーティリティー部品群を提供するオープンソースライブラリです。
authlete-go-gin は [authlete-go][AuthleteGo] ライブラリを使用しており、こちらは
[Authlete Web API][AuthleteAPI] とやりとりするためのオープンソースライブラリです。

この認可サーバーにより発行されたアクセストークンは、Authlete
をバックエンドサービスとして利用しているリソースサーバーに対して使うことができます。
[gin-resource-server][GinResourceServer] はそのようなリソースサーバーの実装です。
[OpenID Connect Core 1.0][OIDCCore]
で定義されている[ユーザー情報エンドポイント][UserInfoEndpoint]をサポートし、
保護リソースエンドポイントの実装例も含んでいます。

ライセンス
----------

  Apache License, Version 2.0

ソースコード
------------

  <code>https://github.com/authlete/gin-oauth-server</code>

Authlete について
-----------------

[Authlete][Authlete] (オースリート) は、OAuth 2.0 & OpenID Connect
の実装をクラウドで提供するサービスです ([概説][AuthleteOverview])。
Authlete が提供するデフォルト実装を使うことにより、もしくはこの実装
(gin-oauth-server) でおこなっているように [Authlete Web API][AuthleteAPI]
を用いて認可サーバーを自分で実装することにより、OAuth 2.0 と OpenID Connect
の機能を簡単に実現できます。

この認可サーバーの実装を使うには、Authlete から API
クレデンシャルズを取得し、`authlete.toml` に設定する必要があります。
API クレデンシャルズを取得する手順はとても簡単です。
単にアカウントを登録するだけで済みます ([サインアップ][AuthleteSignUp])。
詳細は[クイックガイド][AuthleteGettingStarted]を参照してください。


実行方法
--------

1. authlete-go ライブラリと authlete-go-gin ライブラリをインストールします。

        $ go get github.com/authlete/authlete-go
        $ go get github.com/authlete/authlete-go-gin

2. この認可サーバーの実装をダウンロードします。

        $ git clone https://github.com/authlete/gin-oauth-server.git
        $ cd gin-oauth-server

3. 設定ファイルを編集して API クレデンシャルズをセットします。

        $ vi authlete.toml

4. 認可サーバーをビルドします。

        $ make

5. `http://localhost:8080`　で認可サーバーを起動します。

        $ make run

エンドポイント
--------------

この実装は、下表に示すエンドポイントを公開します。

| エンドポイント                     | パス                                |
|:-----------------------------------|:------------------------------------|
| 認可エンドポイント                 | `/api/authorization`                |
| トークンエンドポイント             | `/api/token`                        |
| JWK Set エンドポイント             | `/api/jwks`                         |
| 設定エンドポイント                 | `/.well-known/openid-configuration` |
| 取り消しエンドポイント             | `/api/revocation`                   |
| イントロスペクションエンドポイント | `/api/introspection`                |

認可エンドポイントとトークンエンドポイントは、[RFC 6749][RFC6749]、[OpenID Connect Core 1.0][OIDCCore]、
[OAuth 2.0 Multiple Response Type Encoding Practices][MultiResponseType]、[RFC 7636][RFC7636]
([PKCE][PKCE])、その他の仕様で説明されているパラメーター群を受け付けます。

JWK Set エンドポイントは、クライアントアプリケーションが (1) この OpenID
プロバイダーによる署名を検証できるようにするため、また (2) この OpenID
へのリクエストを暗号化できるようにするため、JSON Web Key Set ドキュメント
(JWK Set) を公開します。

設定エンドポイントは、この OpenID プロバイダーの設定情報を
[OpenID Connect Discovery 1.0][OIDCDiscovery] で定義されている JSON フォーマットで公開します。

取り消しエンドポイントはアクセストークンやリフレッシュトークンを取り消すための
Web API です。 その動作は [RFC 7009][RFC7009] で定義されています。

イントロスペクションエンドポイントはアクセストークンやリフレッシュトークンの情報を取得するための
Web API です。 その動作は [RFC 7662][RFC7662] で定義されています。

認可リクエストの例
------------------

次の例は [Implicit フロー][ImplicitFlow]を用いて認可エンドポイントからアクセストークンを取得する例です。
`{クライアントID}` となっているところは、あなたのクライアントアプリケーションの実際のクライアント
ID で置き換えてください。 クライアントアプリケーションについては、[クイックガイド][AuthleteGettingStarted]
および[開発者コンソール][DeveloperConsole]のドキュメントを参照してください。

    http://localhost:8080/api/authorization?client_id={クライアントID}&response_type=token

上記のリクエストにより、認可ページが表示されます。
認可ページでは、ログイン情報の入力と、"Authorize" ボタン (認可ボタン) もしくは "Deny" ボタン
(拒否ボタン) の押下が求められます。ユーザーデータベースのダミー実装 (`user_management.go`)
は次の二つのアカウントを含んでいます。どちらかを使用してください。

| ログイン ID | パスワード |
|:------------|:-----------|
| `john`      | `john`     |
| `jane`      | `jane`     |

注意
----

- CSRF プロテクションは実装されていません。

その他の情報
------------

- [Authlete][Authlete] - Authlete ホームページ
- [authlete-go][AuthleteGo] - Go 用 Authlete ライブラリ
- [authlete-go-gin][AuthleteGoGin] - Gin (Go) 用 Authlete ライブラリ
- [gin-resource-server][GinResourceServer] - リソースサーバーの実装

コンタクト
----------

コンタクトフォーム : https://www.authlete.com/ja/contact/

| 目的 | メールアドレス       |
|:-----|:---------------------|
| 一般 | info@authlete.com    |
| 営業 | sales@authlete.com   |
| 広報 | pr@authlete.com      |
| 技術 | support@authlete.com |

[Authlete]:               https://www.authlete.com/ja/
[AuthleteAPI]:            https://docs.authlete.com/
[AuthleteGettingStarted]: https://www.authlete.com/ja/developers/getting_started/
[AuthleteOverview]:       https://www.authlete.com/ja/developers/overview/
[AuthleteGo]:             https://github.com/authlete/authlete-go/
[AuthleteGoGin]:          https://github.com/authlete/authlete-go-gin/
[AuthleteSignUp]:         https://so.authlete.com/accounts/signup
[DeveloperConsole]:       https://www.authlete.com/ja/developers/cd_console/
[Gin]:                    https://github.com/gin-gonic/gin
[GinOAuthServer]:         https://github.com/authlete/gin-oauth-server/
[GinResourceServer]:      https://github.com/authlete/gin-resource-server/
[ImplicitFlow]:           https://tools.ietf.org/html/rfc6749#section-4.2
[MultiResponseType]:      https://openid.net/specs/oauth-v2-multiple-response-types-1_0.html
[OIDC]:                   https://openid.net/connect/
[OIDCCore]:               https://openid.net/specs/openid-connect-core-1_0.html
[OIDCDiscovery]:          https://openid.net/specs/openid-connect-discovery-1_0.html
[PKCE]:                   https://www.authlete.com/ja/developers/pkce/
[RFC6749]:                https://tools.ietf.org/html/rfc6749
[RFC7009]:                https://tools.ietf.org/html/rfc7009
[RFC7636]:                https://tools.ietf.org/html/rfc7636
[RFC7662]:                https://tools.ietf.org/html/rfc7662
[UserInfoEndpoint]:       https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
