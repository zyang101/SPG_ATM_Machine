# SPG_ATM_Machine

ATM Machine for Security & Privacy Project

Team Members: Johnson, Justin, Johnny, Reid, and Joshua

**Installing and Setup:**

1. cd into the base directory ~/SPG_ATM_Machine/
2. Run "go mod tidy" to install all the dependencies.
3. Run "go run main" to start program and initialize db if it doesn't exist.
4. (Optional) Download an extension to view SQLite databases for easier data visualization.

**Login Directions:**

1. Enter "go run main" to start program
2. Enter "Y" when prompted to continue
3. Insert "ATM Card" by changing the value in ~/auth/idcard.txt
   * This value can be either admin, customer, or cash handler
   * It is assumed only the respective role will have access to these atm cards.
   * The role value is case sensitive.
4. Enter a valid username (case sensitive) and PIN (6 digits)
5. User is brought to the landing page for their corresponding role.

**Customer Directions:**

1. Upon login, the customer will have the following options (after Login Directions):
   * View the following options again
   * View current account balance
   * Deposit money
   * Withdraw money
   * Transfer funds to another customer
   * Exit the session
2. Follow the on screen instructions to enter amounts for deposits, withdrawals, transfers and to view account balance.
3. Deposits and Withdrawals are must be under the limits set by the admin.
4. All Cash Amounts need to be a valid float to be parsed, no other characters.

**Cash Handler Directions:** 

1. Upon login, the cash handler will have the following options (after Login Directions):

   * View the following options again
   * Get the ATM's cash balance and demonations
   * Deposit money into ATM
   * Withdrawal Money from ATM
   * Exit the session
2. The Deposit and Withdrawal amounts for the Cash Handler are not restricted by the ATM limits
3. All Cash Amounts need to be a valid float to be parsed, no other characters.

**Admin Directions:**

1. Upon login, the admin will have the following options (after Login Directions):
   * View the following options again
   * Create a new customer account
   * View Transaction Histories.
   * Set the ATM deposit and withdrawal limits.
   * Exit the session
2. For account creation, the format must follow the intructions.
   * Usernames are case sensitive
   * Passwords must be a 6 digit pin
   * Name must be alphabetic characters with spaces.
   * Date of birth must be in the for mm/dd/yr
3. All Cash Amounts need to be a valid float to be parsed, no other characters.

**Code File Structure:**

SPG_ATM_Machine/
├── admin/          # Admin role features
├── auth/           # Login & ID verification
├── customer/       # Customer transaction menu
├── handler/        # Cash handler functions
├── internal/       # Database Root Folder
│   ├── api/        # DB queries and core logic
│   └── db/         # SQLite connection
├── utils/          # Input validation & helpers
├── auth/idcard.txt # Role validation file
├── go.mod      #  Imports
└── main.go     # Program to Run
