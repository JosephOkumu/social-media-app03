// Fetch categories from backend
fetch("/categories")
  .then((response) => response.json())
  .then((categories) => {
    const categoriesGrid = document.getElementById("categories-grid");
    categoriesGrid.className = "categories-grid";

    const icons = {
      'Technology': 'fas fa-microchip',
      'Health': 'fas fa-heartbeat',
      'Education': 'fas fa-graduation-cap',
      'Entertainment': 'fas fa-music',
      'Travel': 'fas fa-plane',
      'Food': 'fas fa-utensils',
      'Business': 'fas fa-briefcase',
      'Sports': 'fas fa-basketball-ball',
      'Lifestyle': 'fas fa-cogs',
      'Politics': 'fas fa-landmark'
    };

    categories.forEach((category) => {
      const label = document.createElement("label");
      label.className = "category-checkbox";

      const checkbox = document.createElement("input");
      checkbox.type = "checkbox";
      checkbox.name = "categories[]";
      checkbox.value = category.ID;

      const icon = document.createElement("i");
      icon.className = icons[category.Name] || 'fas fa-tag';

      const span = document.createElement("span");
      span.textContent = category.Name;

      label.appendChild(checkbox);
      label.appendChild(icon);
      label.appendChild(span);
      categoriesGrid.appendChild(label);
    });
  })

class NotificationManager {
  constructor() {
    // Create container if it doesn't exist
    this.container = document.querySelector(".notification-container");
    if (!this.container) {
      this.container = document.createElement("div");
      this.container.className = "notification-container";
      document.body.appendChild(this.container);
    }
  }

  show(message, type = "success", duration = 3000) {
    const notification = document.createElement("div");
    notification.className = `notification ${type}`;

    // Create notification content
    notification.innerHTML = `
            <i class="notification-icon fas ${this.getIcon(type)}"></i>
            <div class="notification-content">${message}</div>
            <button class="notification-close" aria-label="Close notification">×</button>
        `;

    // Add to container
    this.container.appendChild(notification);

    // Setup close button
    const closeBtn = notification.querySelector(".notification-close");
    closeBtn.onclick = () => this.hide(notification);

    // Show with animation
    requestAnimationFrame(() => notification.classList.add("show"));

    // Auto-hide
    if (duration) {
      setTimeout(() => this.hide(notification), duration);
    }
  }

  hide(notification) {
    notification.classList.remove("show");
    setTimeout(() => notification.remove(), 300);
  }

  getIcon(type) {
    switch (type) {
      case "success":
        return "fa-check-circle";
      case "error":
        return "fa-exclamation-circle";
      case "warning":
        return "fa-exclamation-triangle";
      default:
        return "fa-info-circle";
    }
  }
}

// Initialize notification manager
const notificationManager = new NotificationManager();

// Handle form submission
document.querySelector("form").addEventListener("submit", async function (event) {
  event.preventDefault();

  // Validate form inputs
  const title = document.getElementById("title").value.trim();
  const content = document.getElementById("content").value.trim();
  const categories = Array.from(
    document.querySelectorAll('input[name="categories[]"]:checked')
  ).map(checkbox => checkbox.value);
  const image = document.getElementById("image").files[0];
  //check if image size is above 20mbs
  const maxSizeInBytes = 20 * 1024 * 1024;
  if (image && image.size > maxSizeInBytes) {
    notificationManager.show("Image size exceeds the limit of 20MB.", "error");
    return;
  }

  // Input validation
  if (title.length > 50) {
    notificationManager.show("Title exceeds 50 characters.", "error");
    return;
  }
  if (title.length < 2) {
    notificationManager.show("Title is too short.", "error");
    return;
  }
  if (content.length < 5) {
    notificationManager.show("Content is too short.", "error");
    return;
  }
  if (content.length > 500) {
    notificationManager.show("Content exceeds Limit.", "error");
    return;
  }
  if (categories.length === 0) {
    notificationManager.show("Please select at least one category.", "error");
    return;
  }

  try {
    // Handle image upload first if an image is selected
    if (image) {
      const imageData = new FormData();
      imageData.append("image", image);

      const imageResponse = await fetch("/upload-image", {
        method: "POST",
        body: imageData
      });

      if (!imageResponse.ok) {
        const errorData = await imageResponse.json();
        throw new Error(errorData.error || `Image upload failed: ${imageResponse.statusText}`);
      }

      // Wait for the image upload response
      await imageResponse.json();
    }

    // Proceed with creating the post after successful image upload
    const postData = new URLSearchParams();
    postData.append("title", title);
    postData.append("content", content);
    categories.forEach(category => postData.append("categories[]", category));

    const postResponse = await fetch("/create-post", {
      method: "POST",
      body: postData
    });

    if (!postResponse.ok) {
      const errorData = await postResponse.json();
      throw new Error(errorData.message || `Post creation failed: ${postResponse.statusText}`);
    }

    if (postResponse.redirected) {
      notificationManager.show("Post created successfully!", "success");
      // Add a small delay before redirect to show the success message
      setTimeout(() => {
        window.location.href = postResponse.url;
      }, 2000);
    }

  } catch (error) {
    console.error("Error:", error);
    notificationManager.show(
      "Failed to process your request, likely due to invalid image size above 20mbs",
      "error"
    );
  }
});

/**
 * Function to display a notification
 * @param {string} message - The message to display
 * @param {string} type - Type of notification ('success' or 'error')
 */
function showNotification(message, type = "success") {
  const notification = document.createElement("div");
  notification.textContent = message;
  notification.className = `notification ${type}`;

  document.body.appendChild(notification);

  // Show the notification with smooth transitions
  setTimeout(() => {
    notification.classList.add("show");
  }, 10);

  // Hide and remove the notification after 2 seconds
  setTimeout(() => {
    notification.classList.remove("show");
    setTimeout(() => {
      notification.remove();
    }, 500);
  }, 2000);
}

