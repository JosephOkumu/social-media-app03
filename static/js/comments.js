
const commentForm = document.getElementById("comment-form");

commentForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const postID = document.getElementById("view-post").getAttribute("post-id");
    const content = document.getElementById("comment-content").value;

    if (!content.trim()) {
        alert("Comment cannot be empty.");
        return;
    }
    try {
        const response = await fetch(`/comments/create`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                post_id: Number(postID),
                parent_id: null,
                content: content,
            }),
        });

        const text = await response.text();
        if (response.status === 400) {
            alert("Comment cannot be empty.");
            return;
        }
        if (!response.ok || text.startsWith("<")) {
            window.location.href = "/login";
            return;
        }

        const data = JSON.parse(text);
        if (data.status === "success") {
            document.getElementById("comment-content").value = "";
            document.dispatchEvent(new Event("DOMContentLoaded"));
        } else {
            alert("Failed to post comment.");
        }
    } catch (error) {
        console.error("Error posting comment:", error);
        alert("Failed to post comment.");
    }
});

document.addEventListener("DOMContentLoaded", () => {
    const escapeHTML = (str) => {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    };
    const commentCount = document.getElementById("comment-count");
    const viewPostContainer = document.getElementById("view-post");

    if (!viewPostContainer) return;

    const postID = viewPostContainer.getAttribute("post-id");
    const MAX_NESTING_LEVEL = 3;

    const fetchComments = async (postID) => {
        try {
            const response = await fetch(`/comments?post_id=${postID}`);
            return await response.json();
        } catch (error) {
            console.error("Error fetching comments:", error);
            return [];
        }
    };

    const createCommentElement = (commentData, level, postID) => {
        const comment = document.createElement("div");
        comment.classList.add("comment");
        comment.id = `comment-${commentData.id}`;
        comment.dataset.level = level;
        comment.innerHTML = `
            <div class="comment-header">
                <span class="comment-author">${commentData.username}</span>
                <span class="comment-time">${new Date(commentData.created_at).toLocaleString()}</span>
            </div>
            <p class="comment-content">${escapeHTML(commentData.content)}</p>
            <div class="reaction-container">
                <button class="thumbs-up ${commentData.user_reaction === "LIKE" ? "selected" : ""}">
                    <i class="fa-solid fa-thumbs-up"></i> <span>${commentData.likes}</span>
                </button>
                <button class="thumbs-down ${commentData.user_reaction === "DISLIKE" ? "selected" : ""}">
                    <i class="fa-solid fa-thumbs-down"></i> <span>${commentData.dislikes}</span>
                </button>
            </div>
        `;

        if (level < MAX_NESTING_LEVEL) addReplyButton(comment, commentData, postID, level);

        // Add event listeners for thumbs-up and thumbs-down
        const thumbsUpButton = comment.querySelector(".thumbs-up");
        const thumbsDownButton = comment.querySelector(".thumbs-down");

        thumbsUpButton.addEventListener("click", () => handleReaction(commentData.id, "LIKE", thumbsUpButton, thumbsDownButton));
        thumbsDownButton.addEventListener("click", () => handleReaction(commentData.id, "DISLIKE", thumbsDownButton, thumbsUpButton));

        return comment;
    };

    const handleReaction = async (commentID, reactionType, clickedButton, otherButton) => {
        try {
            const response = await fetch("/comments/react", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    comment_id: commentID,
                    reaction_type: reactionType,
                }),
            });

            const text = await response.text(); // Read the response as text first
            if (!response.ok || text.startsWith("<")) {
                // Redirect to login if the response is HTML or not OK
                window.location.href = "/login";
                return;
            }

            const data = JSON.parse(text); // Parse the response as JSON
            const countSpan = clickedButton.querySelector("span");
            const otherCountSpan = otherButton.querySelector("span");

            // Handle reaction based on the backend's response
            if (data.status === "added") {
                // Increment the clicked button's count
                countSpan.textContent = parseInt(countSpan.textContent, 10) + 1;
                clickedButton.classList.add("selected");
            } else if (data.status === "updated") {
                // Update counts: increment the clicked button's count and decrement the other
                countSpan.textContent = parseInt(countSpan.textContent, 10) + 1;
                otherCountSpan.textContent = parseInt(otherCountSpan.textContent, 10) - 1;
                clickedButton.classList.add("selected");
                otherButton.classList.remove("selected");
            } else if (data.status === "removed") {
                // Decrement the clicked button's count
                countSpan.textContent = parseInt(countSpan.textContent, 10) - 1;
                clickedButton.classList.remove("selected");
            }
        } catch (error) {
            console.error("Error reacting to the comment:", error);
            alert("An error occurred. Please try again.");
        }
    };


    const addReplyButton = (comment, commentData, postID, level) => {
        const replyButton = document.createElement("button");
        replyButton.textContent = "Reply";
        replyButton.classList.add("reply-btn");
        replyButton.addEventListener("click", () => toggleReplyForm(comment, commentData, postID));
        comment.appendChild(replyButton);
    };

    const toggleReplyForm = (comment, commentData, postID) => {
        const existingReplyForm = comment.querySelector(".reply-form");
        if (existingReplyForm) {
            existingReplyForm.remove();
        } else {
            const replyForm = document.createElement("div");
            replyForm.classList.add("reply-form");
            replyForm.innerHTML = `
                <textarea placeholder="Write your reply..."></textarea>
                <button class="submit-reply">Submit Reply</button>
                <button class="cancel-reply">Cancel</button>
            `;
            replyForm.querySelector(".cancel-reply").addEventListener("click", () => replyForm.remove());
            replyForm.querySelector(".submit-reply").addEventListener("click", () => submitReply(replyForm, commentData, postID, comment));
            comment.appendChild(replyForm);
        }
    };

    const submitReply = (replyForm, commentData, postID, comment) => {
        const replyContent = replyForm.querySelector("textarea").value;

        if (!replyContent.trim()) {
            alert("Reply cannot be empty.");
            return;
        }

        fetch(`/comments/create`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                post_id: Number(postID),
                parent_id: commentData.id,
                content: replyContent,
            }),
        })
            .then((response) => {
                return response.text().then((text) => {
                    if (!response.ok || text.startsWith("<")) {
                        window.location.href = "/login";
                        return;
                    }
                    return JSON.parse(text);
                });
            })
            .then((data) => {
                if (data && data.status === "success") {
                    const parentLevel = parseInt(comment.dataset.level);
                    const replyLevel = parentLevel + 1;
                    const tempReply = createCommentElement({
                        id: data.id,
                        content: replyContent,
                        username: data.username,
                        created_at: data.created_at,
                        likes: 0,
                        dislikes: 0,
                        user_reaction: null
                    }, replyLevel, postID);
                    let repliesContainer = comment.nextElementSibling;
                    if (!repliesContainer || !repliesContainer.classList.contains("replies-container")) {
                        repliesContainer = document.createElement("div");
                        repliesContainer.classList.add("replies-container");
                        comment.parentElement.insertBefore(repliesContainer, comment.nextSibling);
                    }
                    repliesContainer.style.display = "block";
                    repliesContainer.appendChild(tempReply);

                    // Update reply count display
                    const replyCount = repliesContainer.children.length;
                    let viewRepliesBtn = comment.querySelector(".view-replies-btn");
                    if (!viewRepliesBtn) {
                        viewRepliesBtn = document.createElement("button");
                        viewRepliesBtn.classList.add("view-replies-btn");
                        comment.appendChild(viewRepliesBtn);
                        // Add click event listener
                        viewRepliesBtn.addEventListener("click", () => {
                            const repliesContainer = comment.nextElementSibling;
                            const isHidden = repliesContainer.style.display === "none";
                            repliesContainer.style.display = isHidden ? "block" : "none";
                            viewRepliesBtn.textContent = isHidden ? "Hide Replies" : `+ ${repliesContainer.children.length} Replies`;
                        });
                    }
                    viewRepliesBtn.textContent = "Hide Replies";
                    replyForm.remove();
                } else {
                    alert("Failed to post reply.");
                }
            })
            .catch((error) => {
                console.error("Error posting reply:", error);
                alert("Failed to post reply.");
            });
    };

    const renderReplies = (commentData, repliesContainer, level, postID) => {
        if (commentData.children && commentData.children.length > 0 && level < MAX_NESTING_LEVEL) {
            commentData.children.forEach((childComment) => {
                renderComment(childComment, repliesContainer, level + 1, postID);
            });
        }
    };

    const addViewRepliesButton = (comment, repliesContainer, commentData) => {
        if (commentData.children && commentData.children.length > 0) {
            const viewRepliesButton = document.createElement("button");
            viewRepliesButton.textContent = `+ ${commentData.children.length} Replies`;
            viewRepliesButton.classList.add("view-replies-btn");
            viewRepliesButton.addEventListener("click", () => {
                const isHidden = repliesContainer.style.display === "none";
                repliesContainer.style.display = isHidden ? "block" : "none";
                viewRepliesButton.textContent = isHidden
                    ? "Hide Replies"
                    : `+ ${commentData.children.length} Replies`;
            });
            comment.appendChild(viewRepliesButton);
        }
    };

    const renderComment = (commentData, parentElement, level, postID) => {
        const comment = createCommentElement(commentData, level, postID);
        const repliesContainer = document.createElement("div");
        repliesContainer.classList.add("replies-container");
        repliesContainer.style.display = "none";

        renderReplies(commentData, repliesContainer, level, postID);
        addViewRepliesButton(comment, repliesContainer, commentData);

        parentElement.appendChild(comment);
        parentElement.appendChild(repliesContainer);
    };

    const displayComments = (comments, commentsList, postID) => {
        commentsList.innerHTML = ""; // Clear existing comments
        const totalComments = countComments(comments); // Count all comments, including nested ones
        commentCount.textContent = `${totalComments} comments`;
        comments.forEach((comment) => renderComment(comment, commentsList, 1, postID));
    };

    const init = async () => {
        const commentsList = document.getElementById("comments-list");
        const comments = await fetchComments(postID);
        displayComments(comments, commentsList, postID);
    };

    init();
});

const countComments = (comments) => {
    let count = 0;
    comments.forEach(comment => {
        count += 1; // Count the current comment
        if (comment.children && comment.children.length > 0) {
            count += countComments(comment.children); // Recursively count nested comments
        }
    });
    return count;
};

