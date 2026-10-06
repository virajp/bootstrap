# <%= repo_name %>

Owner: <%= repo_owner %>, repository: <%= repo %>.
<% if merge_model.main == "pr" %>
Reaches main by pull request.
<% endif %>
